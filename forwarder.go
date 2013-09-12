package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"code.google.com/p/go.crypto/ssh"
	"github.com/calmh/mole/configuration"
)

func startForwarder(conn *ssh.ClientConn) chan<- configuration.ForwardLine {
	fwdChan := make(chan configuration.ForwardLine)
	go func() {
		for line := range fwdChan {
			for i := 0; i <= line.Repeat; i++ {
				src := fmt.Sprintf("%s:%d", line.SrcIP, line.SrcPort+i)
				dst := fmt.Sprintf("%s:%d", line.DstIP, line.DstPort+i)

				debug("listen", src)
				l, e := net.Listen("tcp", src)
				if e != nil {
					log.Fatal(e)
				}

				go func(l net.Listener, dst string) {
					for {
						c1, e := l.Accept()
						if e != nil {
							log.Fatal(e)
						}
						debug("accepted", c1.LocalAddr(), c1.RemoteAddr())
						var c2 net.Conn
						debug("dial (ssh)", dst)
						if conn != nil {
							c2, e = conn.Dial("tcp", dst)
						} else {
							c2, e = net.Dial("tcp", dst)
						}
						if e != nil {
							log.Fatal(e)
						}

						go func() {
							n, e := io.Copy(c1, c2)
							if e != nil {
								log.Fatal(e)
							}
							debug("close <-", c1.LocalAddr(), "bytes in:", n)
							c1.Close()
						}()
						go func() {
							n, e := io.Copy(c2, c1)
							if e != nil {
								log.Fatal(e)
							}
							debug("close ->", dst, "bytes out:", n)
							c2.Close()
						}()
					}
				}(l, dst)
			}
		}
	}()
	return fwdChan
}
