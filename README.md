SOCKS
=====

[![GoDoc](https://godoc.org/h12.io/socks?status.svg)](https://godoc.org/h12.io/socks)

SOCKS is a SOCKS4, SOCKS4A and SOCKS5 proxy package for Go.

## Quick Start
### Get the package

    go get -u "github.com/bigemon/socks"

### Import the package

    import "github.com/bigemon/socks"

### Create a SOCKS proxy dialing function

    dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, "127.0.0.1:1080",nil)
    tr := &http.Transport{Dial: dialSocksProxy}
    httpClient := &http.Client{Transport: tr}

## Example

```go
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bigemon/socks"
)

func main() {
	opt := &socks.Opt{
		User:        "admin",
		Password:    "123456",
		DialTimeout: time.Second * 5,
		Timeout:     time.Second * 15,
	}
	//If opt is nil, the default value will be used.       ↓↓
	//socks.DialSocksProxy(socks.SOCKS5, "127.0.0.1:8800", nil) 
	dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, "127.0.0.1:8800", opt)
	tr := &http.Transport{Dial: dialSocksProxy}
	httpClient := &http.Client{Transport: tr}

	bodyText, err := TestHttpsGet(httpClient, "http://2018.ip138.com/ic.asp")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Print(bodyText)
}

func TestHttpsGet(c *http.Client, url string) (bodyText string, err error) {
	resp, err := c.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	bodyText = string(body)
	return
}
```

## Alternatives
http://godoc.org/golang.org/x/net/proxy


