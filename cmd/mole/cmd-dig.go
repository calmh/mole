package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"code.google.com/p/go.crypto/ssh"
	"github.com/jessevdk/go-flags"
	"github.com/sbinet/liner"
	"nym.se/mole/conf"
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

	var cfg *conf.Config
	var err error

	if c.Local {
		cfg, err = conf.LoadFile(args[0])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		cert := certificate()
		cl := NewClient(serverAddr, cert)
		tun := cl.Deobfuscate(cl.Get(args[0]))
		cfg, err = conf.LoadString(tun)
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

	var vpn VPN

	if cfg.Vpnc != nil {
		vpn = startVpn("vpnc", cfg)
	}

	var sshTun *ssh.ClientConn
	if cfg.General.Main != "" {
		sshTun = sshHost(cfg.General.Main, cfg)
		log.Println()
	}

	fwdChan := startForwarder(sshTun)
	sendForwards(fwdChan, cfg)

	shell(fwdChan)

	if vpn != nil {
		vpn.Stop()
	}

	addrs = extraneousAddresses(cfg)
	if len(addrs) > 0 {
		removeAddresses(addrs)
	}

	return nil
}

func shell(fwdChan chan<- conf.ForwardLine) {
	help := func() {
		log.Println("Available commands:")
		log.Println("  help, ?                          - show help")
		log.Println("  quit, ^D                         - stop forwarding and exit")
		log.Println("  debug                            - enable debugging")
		log.Println("  fwd srcip:srcport dstip:dstport  - add forward")
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
		case "help":
			help()
		case "?":
			help()
		case "debug":
			fmt.Println(msgDebugEnabled)
			globalOpts.Debug = true
		case "fwd":
			if len(parts) != 3 {
				log.Printf(msgErrIncorrectFwd, cmd)
				break
			}

			src := strings.SplitN(parts[1], ":", 2)
			if len(src) != 2 {
				log.Printf(msgErrIncorrectFwdSrc, parts[1])
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
				log.Printf(msgErrIncorrectFwdIP, src[0])
				break
			}

			dst := strings.SplitN(parts[2], ":", 2)
			if len(dst) != 2 {
				log.Printf(msgErrIncorrectFwdDst, parts[2])
				break
			}

			srcp, err := strconv.Atoi(src[1])
			if err != nil {
				log.Println(err)
				break
			}
			if srcp < 1024 {
				log.Printf(msgErrIncorrectFwdPriv, srcp)
				break
			}

			dstp, err := strconv.Atoi(dst[1])
			if err != nil {
				log.Println(err)
				break
			}
			fwd := conf.ForwardLine{
				SrcIP:   src[0],
				SrcPort: srcp,
				DstIP:   dst[0],
				DstPort: dstp,
			}
			log.Println("add", fwd)
			fwdChan <- fwd
		default:
			log.Println(msgErrNoSuchCommand, parts[0])
		}

		next <- true
	}
}

func sshHost(host string, cfg *conf.Config) *ssh.ClientConn {
	h := cfg.Hosts[cfg.HostsMap[host]]
	if h.Via != "" {
		cl := sshHost(h.Via, cfg)
		conn, err := cl.Dial("tcp", fmt.Sprintf("%s:%d", h.Addr, h.Port))
		if err != nil {
			log.Fatal(err)
		}
		return sshVia(conn, h)
	} else {
		return sshVia(nil, h)
	}
}

func sendForwards(fwdChan chan<- conf.ForwardLine, cfg *conf.Config) {
	for _, fwd := range cfg.Forwards {
		log.Println(underline(fwd.Name))
		for _, line := range fwd.Lines {
			log.Println("  ", line)
			fwdChan <- line
		}
		log.Println()
	}
}
