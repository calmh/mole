package main

import (
	"fmt"
	"io"
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
		infoln()
		fatalln("dig: missing required option <tunnelname>")
	}

	var cfg *conf.Config
	var err error

	if c.Local {
		cfg, err = conf.LoadFile(args[0])
		if err != nil {
			fatalln(err)
		}
	} else {
		cert := certificate()
		cl := NewClient(serverAddr, cert)
		tun := cl.Deobfuscate(cl.Get(args[0]))
		cfg, err = conf.LoadString(tun)
		fatalErr(err)
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
	} else if cfg.OpenConnect != nil {
		vpn = startVpn("openconnect", cfg)
	}

	var sshTun *ssh.ClientConn
	if cfg.General.Main != "" {
		sshTun = sshHost(cfg.General.Main, cfg)
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
		infoln("Available commands:")
		infoln("  help, ?                          - show help")
		infoln("  quit, ^D                         - stop forwarding and exit")
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

func sshHost(host string, cfg *conf.Config) *ssh.ClientConn {
	h := cfg.Hosts[cfg.HostsMap[host]]
	if h.Via != "" {
		cl := sshHost(h.Via, cfg)
		conn, err := cl.Dial("tcp", fmt.Sprintf("%s:%d", h.Addr, h.Port))
		fatalErr(err)
		return sshVia(conn, h)
	} else {
		return sshVia(nil, h)
	}
}

func sendForwards(fwdChan chan<- conf.ForwardLine, cfg *conf.Config) {
	for _, fwd := range cfg.Forwards {
		infoln(underline(fwd.Name))
		for _, line := range fwd.Lines {
			infoln("  ", line)
			fwdChan <- line
		}
	}
}
