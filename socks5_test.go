package socks

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"testing"
	"time"

	socks5 "github.com/h12w/go-socks5"
	"github.com/phayes/freeport"
)

var httpTestServer = func() *http.Server {
	var err error
	httpTestPort, err := freeport.GetFreePort()
	if err != nil {
		panic(err)
	}
	s := &http.Server{
		Addr: ":" + strconv.Itoa(httpTestPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("hello"))
		}),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go s.ListenAndServe()
	runtime.Gosched()
	tcpReady(httpTestPort, 2*time.Second)
	return s
}()

func newTestSocksServer(withAuth bool) (port int) {
	authenticator := socks5.Authenticator(socks5.NoAuthAuthenticator{})
	if withAuth {
		authenticator = socks5.UserPassAuthenticator{
			Credentials: socks5.StaticCredentials{
				"test_user": "test_pass",
			},
		}
	}
	conf := &socks5.Config{
		Logger: log.New(ioutil.Discard, "", log.LstdFlags),
		AuthMethods: []socks5.Authenticator{
			authenticator,
		},
	}

	srv, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	socksTestPort, err := freeport.GetFreePort()
	if err != nil {
		panic(err)
	}

	go func() {
		if err := srv.ListenAndServe("tcp", "0.0.0.0:"+strconv.Itoa(socksTestPort)); err != nil {
			panic(err)
		}
	}()
	runtime.Gosched()
	tcpReady(socksTestPort, 2*time.Second)
	return socksTestPort
}

func TestSocks5Anonymous(t *testing.T) {
	socksTestPort := newTestSocksServer(false)
	dialSocksProxy := Dial(fmt.Sprintf("socks5://127.0.0.1:%d?timeout=5s", socksTestPort))
	tr := &http.Transport{Dial: dialSocksProxy}
	httpClient := &http.Client{Transport: tr}
	resp, err := httpClient.Get(fmt.Sprintf("http://localhost" + httpTestServer.Addr))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if string(respBody) != "hello" {
		t.Fatalf("expect response hello but got %s", respBody)
	}
}

func TestSocks5Auth(t *testing.T) {
	socksTestPort := newTestSocksServer(true)
	dialSocksProxy := Dial(fmt.Sprintf("socks5://test_user:test_pass@127.0.0.1:%d?timeout=5s", socksTestPort))
	tr := &http.Transport{Dial: dialSocksProxy}
	httpClient := &http.Client{Transport: tr}
	resp, err := httpClient.Get(fmt.Sprintf("http://localhost" + httpTestServer.Addr))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if string(respBody) != "hello" {
		t.Fatalf("expect response hello but got %s", respBody)
	}
}

func tcpReady(port int, timeout time.Duration) {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), timeout)
	if err != nil {
		panic(err)
	}
	conn.Close()
}
