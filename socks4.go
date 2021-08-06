package socks

import (
	"errors"
	"net"
	"time"
)

func (cfg *config) dialSocks4(targetAddr string) (_ net.Conn, err error) {
	socksType := cfg.Proto
	proxy := cfg.Host

	// dial TCP
	conn, err := net.DialTimeout("tcp", proxy, cfg.Timeout)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	// connection request
	host, port, err := splitHostPort(targetAddr)
	if err != nil {
		return nil, err
	}
	ip := net.IPv4(0, 0, 0, 1).To4()
	if socksType == SOCKS4 {
		ip, err = lookupIPv4(host)
		if err != nil {
			return nil, err
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

	resp, err := cfg.sendReceive(conn, req)
	if err != nil {
		return nil, err
	} else if len(resp) != 8 {
		return nil, errors.New("server does not respond properly")
	}
	switch resp[1] {
	case 90:
		// request granted
	case 91:
		return nil, errors.New("socks connection request rejected or failed")
	case 92:
		return nil, errors.New("socks connection request rejected because SOCKS server cannot connect to identd on the client")
	case 93:
		return nil, errors.New("socks connection request rejected because the client program and identd report different user-ids")
	default:
		return nil, errors.New("socks connection request failed, unknown error")
	}
	// clear the deadline before returning
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return nil, err
	}
	return conn, nil
}
