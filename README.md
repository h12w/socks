GOSOCKS
=======

A SOCKS (SOCKS4, SOCKS4A and SOCKS5) Proxy Package for Go

##Quick Start
###Get the package

    go get -u "github.com/hailiang/gosocks"

###Import the package

    import "github.com/hailiang/gosocks"

###Create a SOCKS proxy dialing function

    dialSocksProxy := socks.DialSocksProxy(socks.SOCKS5, "127.0.0.1:1080")
    tr := &http.Transport{Dial: dialSocksProxy}
    httpClient := &http.Client{Transport: tr}

##Complete Documentation
http://go.pkgdoc.org/github.com/hailiang/gosocks

##Alternatives
http://code.google.com/p/go/source/browse/?repo=net#hg%2Fproxy
