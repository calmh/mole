package main

import (
	"fmt"
	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/conf"
	"github.com/calmh/mole/table"
	"github.com/sbinet/liner"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

var errTimeout = fmt.Errorf("connection timeout")

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
	defer term.Close()

	// Receive commands

	commands := make(chan string)
	next := make(chan bool)
	go func() {
		for {
			prompt := "mole> "
			if globalOpts.Debug {
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
			globalOpts.Debug = true
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
			fwd := conf.ForwardLine{
				SrcIP:   src[0],
				SrcPort: srcp,
				DstIP:   dst[0],
				DstPort: dstp,
			}
			okln("add", fwd)
			fwdChan <- fwd
		default:
			warnln(msgErrNoSuchCommand, parts[0])
		}

		next <- true
	}
}

func printStats() {
	var rows [][]string
	rows = append(rows, []string{"FORWARD", "CONNS", "IN", "OUT"})
	total := trafficCounter{name: "Total"}
	for _, cnt := range globalConnectionStats {
		rows = append(rows, cnt.row())
		total.conns += cnt.conns
		total.in += cnt.in
		total.out += cnt.out
	}
	rows = append(rows, total.row())
	fmt.Println(table.Fmt("lrrr", rows))
}

func printTotalStats() {
	total := trafficCounter{}
	for _, cnt := range globalConnectionStats {
		total.conns += cnt.conns
		total.in += cnt.in
		total.out += cnt.out
	}
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

	var scanWg sync.WaitGroup
	scanWg.Add(len(cfg.Forwards))
	go func() {
		for _, fwd := range cfg.Forwards {
			go func(fwd conf.Forward) {
				res := forwardTest{name: fwd.Name}
				for _, line := range fwd.Lines {
					for i := 0; i <= line.Repeat; i++ {
						subres := make(chan error, 1)

						go func() {
							time.Sleep(5 * time.Second)
							subres <- errTimeout
						}()

						t0 := time.Now()
						go func() {
							conn, err := dialer.Dial("tcp", line.DstString(i))
							if conn != nil {
								defer conn.Close()
							}
							subres <- err
						}()
						err := <-subres
						ms := time.Since(t0).Seconds() * 1000

						res.results = append(res.results, testResult{dst: line.DstString(i), ms: ms, err: err})
					}
				}
				results <- res
				scanWg.Done()
			}(fwd)
		}
		scanWg.Wait()
		close(results)
	}()

	return results
}
