package main

import (
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/sbinet/liner"
	"io"
	"log"
	"net"
	"os"

	"github.com/calmh/mole/configuration"
	"github.com/jessevdk/go-flags"
)

type cmdDig struct {
	Local bool `short:"l" long:"local" description:"Local file, not remote tunnel definition"`
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
		cl := NewClient(serverAddr, cert)
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
	log.Println()
	forwards(client, cfg)

	shell()

	addrs = extraneousAddresses(cfg)
	if len(addrs) > 0 {
		removeAddresses(addrs)
	}

	log.Println(bold(green("ok")))
	return nil
}

func shell() {
	help := func() {
		log.Println("Available commands:")
		log.Println("  help - shows help")
		log.Println("  quit - stops forwarding and exits")
		log.Println("  debug - enable debugging")
	}

	term := liner.NewLiner()
loop:
	for {
		cmd, err := term.Prompt("mole> ")
		if err == io.EOF {
			fmt.Println()
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		switch cmd {
		case "":
		case "quit":
			break loop
		case "help":
			help()
		case "?":
			help()
		case "debug":
			fmt.Println("debug output enabled")
			globalOpts.Debug = true
		default:
			log.Println("what? \"help\" might help.")
		}
	}
	term.Close()
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

				if globalOpts.Debug {
					log.Println("listen", src)
				}
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
						if globalOpts.Debug {
							log.Println("accepted", c1.LocalAddr(), c1.RemoteAddr())
						}

						if globalOpts.Debug {
							log.Println("dial", dst)
						}
						c2, e := conn.Dial("tcp", dst)
						if e != nil {
							panic(e)
						}

						go func() {
							n, e := io.Copy(c1, c2)
							if e != nil {
								log.Fatal(e)
							}
							if globalOpts.Debug {
								log.Println("close <-", c1.LocalAddr(), "bytes in:", n)
							}
							c1.Close()
						}()
						go func() {
							n, e := io.Copy(c2, c1)
							if e != nil {
								log.Fatal(e)
							}
							if globalOpts.Debug {
								log.Println("close ->", dst, "bytes out:", n)
							}
							c2.Close()
						}()
					}
				}(l, dst)
			}
		}
		log.Println()
	}
}
