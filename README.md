SOCKS
=====

[![GoDoc](https://godoc.org/h12.io/socks?status.svg)](https://godoc.org/h12.io/socks)

A SOCKS is a SOCKS4, SOCKS4A and SOCKS5 proxy package for Go.

The package provides `socks.Dial` which returns a TCP dialing function from a socks proxy connection string. The returned dialing function can then be used to establish a TCP connection via the socks proxy or be used to initialize `http.Transport` for an HTTP connection.

## Quick Start
### Get the package

    go get -u "h12.io/socks"

### Import the package

    import "h12.io/socks"

### Create a SOCKS proxy dialing function

    dialSocksProxy := socks.Dial("socks5://127.0.0.1:1080?timeout=5s")
    tr := &http.Transport{Dial: dialSocksProxy}
    httpClient := &http.Client{Transport: tr}

### User/password authentication

    dialSocksProxy := socks.Dial("socks5://user:password@127.0.0.1:1080?timeout=5s")

## Example

```go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"h12.io/socks"
)

func main() {
	dialSocksProxy := socks.Dial("socks5://127.0.0.1:1080?timeout=5s")
	tr := &http.Transport{Dial: dialSocksProxy}
	httpClient := &http.Client{Transport: tr}
	resp, err := httpClient.Get("http://www.google.com")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.StatusCode)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(buf))
}
```
