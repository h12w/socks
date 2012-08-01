gosocks
=======

A SOCKS (SOCKS4, SOCKS4A and SOCKS5) Proxy Package for Go

A simple example:

    dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, "127.0.0.1:1080")
    tr := &http.Transport{Dial: dialSocksProxy}
    httpClient := &http.Client{Transport: tr}

A complete documentation:

http://go.pkgdoc.org/github.com/hailiang/gosocks
