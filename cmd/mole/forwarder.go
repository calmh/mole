package main

import (
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

				debugln("listen", src)
				l, e := net.Listen("tcp", src)
				fatalErr(e)

				go func(l net.Listener, dst string) {
					for {
						c1, e := l.Accept()
						fatalErr(e)
						debugln("accepted", c1.LocalAddr(), c1.RemoteAddr())
						var c2 net.Conn
						if conn != nil {
							debugln("dial (ssh)", dst)
							c2, e = conn.Dial("tcp", dst)
						} else {
							debugln("dial (direct)", dst)
							c2, e = net.Dial("tcp", dst)
						}
						if e != nil {
							// Connection problems here are not fatal; just log them.
							warnln(e)
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
			debugln("close (r)")
			src.Close()
			dst.Close()
			break
		}

		_, e = dst.Write(buf[:n])

		if e != nil {
			debugln("close (w)")
			src.Close()
			dst.Close()
			break
		}
	}
}
