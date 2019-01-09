// Copyright 2012, Hailiang Wang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package socks implements a SOCKS (SOCKS4, SOCKS4A and SOCKS5) proxy client.

A complete example using this package:
	package main

	import (
		"h12.io/socks"
		"fmt"
		"net/http"
		"io/ioutil"
	)

	func main() {
		dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, "127.0.0.1:1080",nil)
		tr := &http.Transport{Dial: dialSocksProxy}
		httpClient := &http.Client{Transport: tr}

		bodyText, err := TestHttpsGet(httpClient, "https://h12.io/about")
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Print(bodyText)
	}

	func TestHttpsGet(c *http.Client, url string) (bodyText string, err error) {
		resp, err := c.Get(url)
		if err != nil { return }
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil { return }
		bodyText = string(body)
		return
	}
*/
package socks

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

// Constants to choose which version of SOCKS protocol to use.
const (
	SOCKS4 = iota
	SOCKS4A
	SOCKS5
)

//Opt Some optional parameters
type Opt struct {
	User        string
	Password    string
	DialTimeout time.Duration //default : 5 seconds
	Timeout     time.Duration //default : 15 seconds
}

// DialSocksProxy returns the dial function to be used in http.Transport object.
// Argument socksType should be one of SOCKS4, SOCKS4A and SOCKS5.
// Argument proxy should be in this format "127.0.0.1:1080".
func DialSocksProxy(socksType int, proxy string, opt ...*Opt) func(string, string) (net.Conn, error) {
	var o *Opt
	if len(opt) == 0 {
		o = new(Opt)
	} else {
		o = opt[0]
	}
	if o.Timeout == 0 {
		o.Timeout = time.Second * 15
	}
	if o.DialTimeout == 0 {
		o.DialTimeout = o.Timeout
	}
	if socksType == SOCKS5 {
		return func(_, targetAddr string) (conn net.Conn, err error) {
			return dialSocks5(proxy, targetAddr, o)
		}
	}

	// SOCKS4, SOCKS4A
	return func(_, targetAddr string) (conn net.Conn, err error) {
		return dialSocks4(socksType, proxy, targetAddr, o)
	}
}

func dialSocks5(proxy, targetAddr string, opt *Opt) (conn net.Conn, err error) {
	// dial TCP
	conn, err = net.DialTimeout("tcp", proxy, opt.DialTimeout)
	if err != nil {
		return
	}

	// version identifier/method selection request
	var req []byte
	if opt.User == "" {
		req = []byte{
			5, // version number
			1, // number of methods
			0, // method 0: no authentication
		}
	} else {
		req = []byte{
			5, // version number
			2, // number of methods
			0, // method 0: no authentication
			2, //method 2: user:password authentication
		}
	}
	resp, err := sendReceive(conn, req, opt.Timeout)
	if err != nil {
		return
	} else if len(resp) != 2 {
		err = errors.New("server does not respond properly")
		return
	} else if resp[0] != 5 {
		err = errors.New("server does not support Socks 5")
		return
	} else if resp[1] != 0 && (opt.User != "" && resp[1] != 2) { // no auth (0)  /  user:password (2)
		err = errors.New("socks method negotiation failed")
		return
	}
	//auth user:password info
	if resp[1] == 2 {
		req = make([]byte, 0, 255)
		req = []byte{
			1, //version number
			byte(len(opt.User)),
		}
		req = append(req, []byte(opt.User)...)
		req = append(req, byte(len(opt.Password)))
		req = append(req, []byte(opt.Password)...)
		resp, err = sendReceive(conn, req, opt.Timeout)
		if err != nil {
			return
		} else if len(resp) != 2 {
			err = errors.New("server does not respond properly")
			return
		} else if resp[1] != 0 {
			err = errors.New("authentication failed")
			return
		}
	}

	// detail request
	host, port, err := splitHostPort(targetAddr)
	req = []byte{
		5,               // version number
		1,               // connect command
		0,               // reserved, must be zero
		3,               // address type, 3 means domain name
		byte(len(host)), // address length
	}
	req = append(req, []byte(host)...)
	req = append(req, []byte{
		byte(port >> 8), // higher byte of destination port
		byte(port),      // lower byte of destination port (big endian)
	}...)
	resp, err = sendReceive(conn, req, opt.Timeout)
	if err != nil {
		return
	} else if len(resp) != 10 {
		err = errors.New("Server does not respond properly.")
	} else if resp[1] != 0 {
		err = errors.New("Can't complete SOCKS5 connection.")
	}

	return
}

func dialSocks4(socksType int, proxy, targetAddr string, opt *Opt) (conn net.Conn, err error) {
	// dial TCP
	conn, err = net.DialTimeout("tcp", proxy, opt.DialTimeout)
	if err != nil {
		return
	}

	// connection request
	host, port, err := splitHostPort(targetAddr)
	if err != nil {
		return
	}
	ip := net.IPv4(0, 0, 0, 1).To4()
	if socksType == SOCKS4 {
		ip, err = lookupIP(host)
		if err != nil {
			return
		}
	}
	req := []byte{
		4,                          // version number
		1,                          // command CONNECT
		byte(port >> 8),            // higher byte of destination port
		byte(port),                 // lower byte of destination port (big endian)
		ip[0], ip[1], ip[2], ip[3], // special invalid IP address to indicate the host name is provided
		0, // user id is empty, anonymous proxy only
	}
	if socksType == SOCKS4A {
		req = append(req, []byte(host+"\x00")...)
	}

	resp, err := sendReceive(conn, req, opt.Timeout)
	if err != nil {
		return
	} else if len(resp) != 8 {
		err = errors.New("Server does not respond properly.")
		return
	}
	switch resp[1] {
	case 90:
		// request granted
	case 91:
		err = errors.New("Socks connection request rejected or failed.")
	case 92:
		err = errors.New("Socks connection request rejected becasue SOCKS server cannot connect to identd on the client.")
	case 93:
		err = errors.New("Socks connection request rejected because the client program and identd report different user-ids.")
	default:
		err = errors.New("Socks connection request failed, unknown error.")
	}
	return
}

func sendReceive(conn net.Conn, req []byte, timeout time.Duration) (resp []byte, err error) {
	conn.SetDeadline(time.Now().Add(timeout))
	defer conn.SetDeadline(time.Time{})
	_, err = conn.Write(req)
	if err != nil {
		return
	}
	resp, err = readAll(conn)
	return
}

func readAll(conn net.Conn) (resp []byte, err error) {
	resp = make([]byte, 1024)
	n, err := conn.Read(resp)
	resp = resp[:n]
	return
}

func lookupIP(host string) (ip net.IP, err error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return
	}
	if len(ips) == 0 {
		err = errors.New(fmt.Sprintf("Cannot resolve host: %s.", host))
		return
	}
	ip = ips[0].To4()
	if len(ip) != net.IPv4len {
		fmt.Println(len(ip), ip)
		err = errors.New("IPv6 is not supported by SOCKS4.")
		return
	}
	return
}

func splitHostPort(addr string) (host string, port uint16, err error) {
	host, portStr, err := net.SplitHostPort(addr)
	portInt, err := strconv.ParseUint(portStr, 10, 16)
	port = uint16(portInt)
	return
}
