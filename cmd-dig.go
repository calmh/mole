package main

import (
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/calmh/mole/configuration"
	"github.com/jessevdk/go-flags"
)

type cmdDig struct {
	Local     bool `short:"l" long:"local" description:"Local file, not remote tunnel definition"`
	NoSpinner bool `short:"n" long:"no-blinkenlights" description:"Do not display awesome blinkelights"`
}

var digParser *flags.Parser

func init() {
	cmd := cmdDig{}
	digParser = globalParser.AddCommand("dig", "Dig a tunnel", "'dig' connects to a remote destination and sets up configured local TCP tunnels", &cmd)
}

func (c *cmdDig) Usage() string {
	return "<tunnelname> [dig-OPTIONS]"
}

func (c *cmdDig) Execute(args []string) error {
	setup()

	if len(args) != 1 {
		digParser.WriteHelp(os.Stdout)
		fmt.Println()
		return fmt.Errorf("dig: missing required option <tunnelname>\n")
	}

	var cfg *configuration.Config
	var err error

	if c.Local {
		cfg, err = configuration.LoadFile(args[0])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		cert := certificate()
		cl := NewClient("mole.nym.se:9443", cert)
		cfg, err = configuration.LoadString(cl.Get(args[0]))
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg == nil {
		return fmt.Errorf("no tunnel loaded")
	}

	addrs := missingAddresses(cfg)
	if len(addrs) > 0 {
		addAddresses(addrs)
	}

	client := sshHost(cfg.General.Main, cfg)
	log.Println(bold(green("    Connected.")))
	log.Println()
	forwards(client, cfg)

	log.Println(bold("^C"), "to quit")
	log.Println()

	stopSpinner := make(chan bool)
	if !globalOpts.Debug && !c.NoSpinner {
		go func() {
			fmt.Printf(ansiHideCursor)
		loop:
			for {
				select {
				case <-stopSpinner:
					break loop
				default:
					fmt.Printf("%s%c%c%c%c%c%s\r", ansiFgYellow, 0x2800+rand.Intn(0x80), 0x2800+rand.Intn(0x80), 0x2800+rand.Intn(0x80), 0x2800+rand.Intn(0x80), 0x2800+rand.Intn(0x80), ansiFgReset)
					time.Sleep(250 * time.Millisecond)
				}
			}
			fmt.Printf(ansiKillLine)
			fmt.Printf(ansiShowCursor)
			stopSpinner <- true
		}()
	}

	sigrecv := make(chan os.Signal, 1)
	signal.Notify(sigrecv, os.Interrupt, syscall.SIGTERM)
	<-sigrecv

	if !globalOpts.Debug && !c.NoSpinner {
		stopSpinner <- true
		<-stopSpinner
		fmt.Println()
	}

	addrs = extraneousAddresses(cfg)
	if len(addrs) > 0 {
		removeAddresses(addrs)
	}

	return nil
}

func sshHost(host string, cfg *configuration.Config) *ssh.ClientConn {
	h := cfg.Hosts[cfg.HostsMap[host]]
	if h.Via != "" {
		cl := sshHost(h.Via, cfg)
		conn, err := cl.Dial("tcp", fmt.Sprintf("%s:%d", h.Addr, h.Port))
		if err != nil {
			panic(err)
		}
		return sshVia(conn, h)
	} else {
		return sshVia(nil, h)
	}
}

func forwards(conn *ssh.ClientConn, cfg *configuration.Config) {
	for _, fwd := range cfg.Forwards {
		log.Println(underline(fwd.Name))
		for _, line := range fwd.Lines {
			if line.Repeat == 0 {
				src := fmt.Sprintf(cyan("%s:%d"), line.SrcIP, line.SrcPort)
				dst := fmt.Sprintf("%s:%d", line.DstIP, line.DstPort)
				log.Printf("  %-37s -> %s", src, dst)
			} else {
				src := fmt.Sprintf(cyan("%s:%d-%d"), line.SrcIP, line.SrcPort, line.SrcPort+line.Repeat)
				dst := fmt.Sprintf("%s:%d-%d", line.DstIP, line.DstPort, line.DstPort+line.Repeat)
				log.Printf("  %-37s -> %s", src, dst)
			}
			for i := 0; i <= line.Repeat; i++ {
				src := fmt.Sprintf("%s:%d", line.SrcIP, line.SrcPort+i)
				dst := fmt.Sprintf("%s:%d", line.DstIP, line.DstPort+i)

				l, e := net.Listen("tcp", src)
				if e != nil {
					panic(e)
				}

				go func(l net.Listener, dest string) {
					for {
						c1, e := l.Accept()
						if e != nil {
							panic(e)
						}

						c2, e := conn.Dial("tcp", dest)
						if e != nil {
							panic(e)
						}
						go io.Copy(c1, c2)
						go io.Copy(c2, c1)
					}
				}(l, dst)
			}
		}
		log.Println()
	}
}
