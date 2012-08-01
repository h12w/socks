gosocks
=======

A SOCKS (SOCKS4, SOCKS4A and SOCKS5) Proxy Package for Go

Quick start:
1. Get the package:

    go get -u "github.com/hailiang/gosocks"

2. Import the package:

    import "github.com/hailiang/gosocks"

3. Create a customized dial function for the Transport object:

    dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, "127.0.0.1:1080")
    tr := &http.Transport{Dial: dialSocksProxy}
    httpClient := &http.Client{Transport: tr}

A complete documentation:  http://go.pkgdoc.org/github.com/hailiang/gosocks
