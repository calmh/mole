package main

import (
	"log"
	"net"
	"sync/atomic"

	"code.google.com/p/go.crypto/ssh"
	"nym.se/mole/conf"
)

func startForwarder(conn *ssh.ClientConn) chan<- conf.ForwardLine {
	fwdChan := make(chan conf.ForwardLine)
	go func() {
		for line := range fwdChan {
			for i := 0; i <= line.Repeat; i++ {
				src := line.SrcString(i)
				dst := line.DstString(i)

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
						if conn != nil {
							debug("dial (ssh)", dst)
							c2, e = conn.Dial("tcp", dst)
						} else {
							debug("dial (direct)", dst)
							c2, e = net.Dial("tcp", dst)
						}
						if e != nil {
							// Connection problems here are not fatal; just log them.
							log.Println(e)
							c1.Close()
							continue
						}

						go copyData(c1, c2, &globalStats.dataIn)
						go copyData(c2, c1, &globalStats.dataOut)
					}
				}(l, dst)
			}
		}
	}()
	return fwdChan
}

func copyData(dst net.Conn, src net.Conn, counter *uint64) {
	buf := make([]byte, 10240)
	for {
		n, e := src.Read(buf)
		atomic.AddUint64(counter, uint64(n))

		if e != nil {
			debug("close (r)")
			src.Close()
			dst.Close()
			break
		}

		_, e = dst.Write(buf[:n])

		if e != nil {
			debug("close (w)")
			src.Close()
			dst.Close()
			break
		}
	}
}
