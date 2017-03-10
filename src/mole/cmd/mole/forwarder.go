package main

import (
	"fmt"
	"mole/conf"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type trafficCounter struct {
	name  string
	conns uint64
	in    uint64
	out   uint64
}

var (
	globalConnectionStats     []*trafficCounter
	globalConnectionStatsLock sync.Mutex
)

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
			for i := 0; i < len(line.Src.Ports); i++ {
				src := line.SrcString(i)
				dst := line.DstString(i)

				debugln("listen", src)
				l, e := net.Listen("tcp", src)
				fatalErr(e)

				cnt := trafficCounter{name: dst}
				globalConnectionStatsLock.Lock()
				globalConnectionStats = append(globalConnectionStats, &cnt)
				globalConnectionStatsLock.Unlock()

				go func(l net.Listener, dst string, cnt *trafficCounter) {
					for {
						c1, e := l.Accept()
						fatalErr(e)
						debugln("accepted", c1.LocalAddr(), c1.RemoteAddr())
						var c2 net.Conn
						t0 := time.Now()
						debugln("dial", dst)
						c2, e = dialer.Dial("tcp", dst)
						if e != nil {
							// Connection problems here are not fatal; just log them.
							warnln(e)
							_ = c1.Close()
							continue
						}
						debugf("dial %s complete in %.01f ms", dst, time.Since(t0).Seconds()*1000)

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
	n, _ := io.Copy(dst, src)
	atomic.AddUint64(counter, uint64(n))
	_ = src.Close()
	_ = dst.Close()
}

func formatBytes(n uint64) string {
	if n < 1024 {
		return fmt.Sprintf("%d ", n)
	}

	prefixes := []string{" k", " M", " G", " T"}
	divisor := 1024.0
	for i := range prefixes {
		rem := float64(n) / divisor
		if rem < 1024.0 || i == len(prefixes)-1 {
			return fmt.Sprintf("%.02f%s", rem, prefixes[i])
		}
		divisor *= 1024
	}
	return ""
}
