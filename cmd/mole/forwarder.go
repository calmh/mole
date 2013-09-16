package main

import (
	"fmt"
	"net"
	"sync/atomic"

	"nym.se/mole/conf"
)

type trafficCounter struct {
	name  string
	conns uint64
	in    uint64
	out   uint64
}

var globalConnectionStats []*trafficCounter

func (cnt trafficCounter) row() []string {
	return []string{cnt.name, fmt.Sprintf("%d", cnt.conns), formatBytes(cnt.in) + "B", formatBytes(cnt.out) + "B"}
}

type Dialer interface {
	Dial(network, addr string) (c net.Conn, err error)
}

func startForwarder(dialer Dialer) chan<- conf.ForwardLine {
	fwdChan := make(chan conf.ForwardLine)
	go func() {
		for line := range fwdChan {
			for i := 0; i <= line.Repeat; i++ {
				src := line.SrcString(i)
				dst := line.DstString(i)

				debugln("listen", src)
				l, e := net.Listen("tcp", src)
				fatalErr(e)

				cnt := trafficCounter{name: dst}
				globalConnectionStats = append(globalConnectionStats, &cnt)

				go func(l net.Listener, dst string, cnt *trafficCounter) {
					for {
						c1, e := l.Accept()
						fatalErr(e)
						debugln("accepted", c1.LocalAddr(), c1.RemoteAddr())
						var c2 net.Conn
						debugln("dial", dst)
						c2, e = dialer.Dial("tcp", dst)
						if e != nil {
							// Connection problems here are not fatal; just log them.
							warnln(e)
							_ = c1.Close()
							continue
						}

						atomic.AddUint64(&cnt.conns, 1)
						go copyData(c1, c2, &cnt.in)
						go copyData(c2, c1, &cnt.out)
					}
				}(l, dst, &cnt)
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
			_ = src.Close()
			_ = dst.Close()
			break
		}

		_, e = dst.Write(buf[:n])

		if e != nil {
			debugln("close (w)")
			_ = src.Close()
			_ = dst.Close()
			break
		}
	}
}
