package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/conf"
	"github.com/calmh/mole/table"
	"github.com/sbinet/liner"
)

var errTimeout = fmt.Errorf("connection timeout")

const maxOutstandingTests = 16 // max number of parallell connection attempts when performing test

func shell(fwdChan chan<- conf.ForwardLine, cfg *conf.Config, dialer Dialer) {
	help := func() {
		infoln("Available commands:")
		infoln("  help, ?                          - show help")
		infoln("  quit, ^D                         - stop forwarding and exit")
		infoln("  test                             - test each forward for connection")
		infoln("  stat                             - show forwarding statistics")
		infoln("  debug                            - enable debugging")
		infoln("  fwd srcip:srcport dstip:dstport  - add forward")
	}

	term := liner.NewLiner()
	atExit(func() {
		term.Close()
	})

	// Receive commands

	commands := make(chan string)
	next := make(chan bool)
	go func() {
		for {
			prompt := "mole> "
			if debugEnabled {
				prompt = "(debug) mole> "
			}
			cmd, err := term.Prompt(prompt)
			if err == io.EOF {
				fmt.Println("quit")
				commands <- "quit"
				return
			}

			if cmd != "" {
				commands <- cmd
				term.AppendHistory(cmd)
				_, ok := <-next
				if !ok {
					return
				}
			}
		}
	}()

	// Catch ^C and treat as "quit" command

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		<-sigchan
		fmt.Println("quit")
		commands <- "quit"
	}()

	// Handle commands

	for {
		cmd := <-commands

		parts := strings.SplitN(cmd, " ", -1)

		switch parts[0] {
		case "quit":
			close(next)
			return
		case "help", "?":
			help()
		case "stat":
			printStats()
		case "test":
			results := testForwards(dialer, cfg)
			for res := range results {
				infof(ansi.Bold(ansi.Cyan(res.name)))
				for _, line := range res.results {
					if line.err == nil {
						infof("%22s %s in %.02f ms", line.dst, ansi.Bold(ansi.Green("-ok-")), line.ms)
					} else {
						infof("%22s %s in %.02f ms (%s)", line.dst, ansi.Bold(ansi.Red("fail")), line.ms, line.err)
					}
				}
			}
		case "debug":
			infoln(msgDebugEnabled)
			debugEnabled = true
		case "fwd":
			if len(parts) != 3 {
				warnf(msgErrIncorrectFwd, cmd)
				break
			}

			src := strings.SplitN(parts[1], ":", 2)
			if len(src) != 2 {
				warnf(msgErrIncorrectFwdSrc, parts[1])
				break
			}

			var ipExists bool
			for _, ip := range currentAddresses() {
				if ip == src[0] {
					ipExists = true
					break
				}
			}
			if !ipExists {
				warnf(msgErrIncorrectFwdIP, src[0])
				break
			}

			dst := strings.SplitN(parts[2], ":", 2)
			if len(dst) != 2 {
				warnf(msgErrIncorrectFwdDst, parts[2])
				break
			}

			srcp, err := strconv.Atoi(src[1])
			if err != nil {
				warnln(err)
				break
			}
			if srcp < 1024 {
				warnf(msgErrIncorrectFwdPriv, srcp)
				break
			}

			dstp, err := strconv.Atoi(dst[1])
			if err != nil {
				warnln(err)
				break
			}
			srcpa := conf.Addrports{
				Addr:  net.ParseIP(src[0]),
				Ports: []int{srcp},
			}
			dstpa := conf.Addrports{
				Addr:  net.ParseIP(dst[0]),
				Ports: []int{dstp},
			}
			fwd := conf.ForwardLine{
				Src: srcpa,
				Dst: dstpa,
			}
			okln("add", fwd)
			fwdChan <- fwd
		default:
			warnf(msgErrNoSuchCommand, parts[0])
		}

		next <- true
	}
}

func printStats() {
	var rows [][]string
	rows = append(rows, []string{"FORWARD", "CONNS", "IN", "OUT"})
	total := trafficCounter{name: "Total"}
	globalConnectionStatsLock.Lock()
	for _, cnt := range globalConnectionStats {
		rows = append(rows, cnt.row())
		total.conns += cnt.conns
		total.in += cnt.in
		total.out += cnt.out
	}
	globalConnectionStatsLock.Unlock()
	rows = append(rows, total.row())
	fmt.Println(table.Fmt("lrrr", rows))
}

func printTotalStats() {
	total := trafficCounter{}
	globalConnectionStatsLock.Lock()
	for _, cnt := range globalConnectionStats {
		total.conns += cnt.conns
		total.in += cnt.in
		total.out += cnt.out
	}
	globalConnectionStatsLock.Unlock()
	if total.conns > 0 {
		infof("Total: %d connections, %sB in, %sB out", total.conns, formatBytes(total.in), formatBytes(total.out))
	}
}

type forwardTest struct {
	name    string
	results []testResult
}
type testResult struct {
	dst string
	ms  float64
	err error
}

func testForwards(dialer Dialer, cfg *conf.Config) <-chan forwardTest {
	results := make(chan forwardTest)
	outstanding := make(chan bool, maxOutstandingTests)

	// Do the test in the background
	go func() {
		var scanWg sync.WaitGroup
		scanWg.Add(len(cfg.Forwards))

		for _, fwd := range cfg.Forwards {
			// Do each forward in parallell
			go func(fwd conf.Forward) {
				nlines := 0
				for _, line := range fwd.Lines {
					nlines += len(line.Src.Ports)
				}

				res := forwardTest{name: fwd.Name}
				res.results = make([]testResult, nlines)

				var fwdWg sync.WaitGroup
				fwdWg.Add(nlines)
				j := 0

				for _, line := range fwd.Lines {
					for i := 0; i < len(line.Src.Ports); i++ {
						// Do each line in parallell
						go func(line conf.ForwardLine, i, j int) {
							outstanding <- true
							t0 := time.Now()
							err := <-testLineIndex(dialer, line, i)
							ms := time.Since(t0).Seconds() * 1000
							<-outstanding

							res.results[j] = testResult{dst: line.DstString(i), ms: ms, err: err}
							fwdWg.Done()
						}(line, i, j)
						j++
					}
				}

				fwdWg.Wait()
				results <- res
				scanWg.Done()
			}(fwd)
		}

		scanWg.Wait()
		close(results)
	}()

	return results
}

func testLineIndex(dialer Dialer, line conf.ForwardLine, i int) <-chan error {
	subres := make(chan error, 1)

	go func() {
		time.Sleep(5 * time.Second)
		subres <- errTimeout
	}()

	go func() {
		debugln("test, Src:", line.SrcString(i), " Dst:", line.DstString(i))
		conn, err := dialer.Dial("tcp", line.DstString(i))
		if err == nil && conn != nil {
			conn.Close()
		}
		subres <- err
	}()

	return subres
}
