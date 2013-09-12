package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"code.google.com/p/go.crypto/ssh"
	"github.com/calmh/mole/configuration"
	"github.com/jessevdk/go-flags"
	"github.com/sbinet/liner"
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
		tun := cl.Deobfuscate(cl.Get(args[0]))
		cfg, err = configuration.LoadString(tun)
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

	var vpn *Vpnc
	if cfg.Vpnc != nil {
		vpn = vpnc(cfg)
	}

	var sshTun *ssh.ClientConn
	if cfg.General.Main != "" {
		sshTun = sshHost(cfg.General.Main, cfg)
		log.Println()
	}

	fwdChan := startForwarder(sshTun)
	sendForwards(fwdChan, cfg)

	shell()

	if vpn != nil {
		vpn.Stop()
	}

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
		log.Println("  help, ?   - show help")
		log.Println("  quit, ^D  - stop forwarding and exit")
		log.Println("  debug     - enable debugging")
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
				prompt = "(debug) mole>"
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

		switch cmd {
		case "quit":
			close(next)
			return
		case "help":
			help()
		case "?":
			help()
		case "debug":
			fmt.Println("debug output enabled")
			globalOpts.Debug = true
		default:
			log.Println("what? try \"help\"")
		}

		next <- true
	}
}

func sshHost(host string, cfg *configuration.Config) *ssh.ClientConn {
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

func sendForwards(fwdChan chan<- configuration.ForwardLine, cfg *configuration.Config) {
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
			fwdChan <- line
		}
		log.Println()
	}
}
