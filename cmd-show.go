package main

import (
	"fmt"
	"github.com/calmh/mole/configuration"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

type cmdShow struct {
	Raw bool `short:"r" long:"raw" description:"Show raw, not parsed, tunnel file"`
}

var showParser *flags.Parser

func init() {
	cmd := cmdShow{}
	showParser = globalParser.AddCommand("show", "Show a tunnel", "'show' show the tunnel definition file without connecting to it", &cmd)
}

func (c *cmdShow) Usage() string {
	return "<tunnelname> [show-OPTIONS]"
}

func (c *cmdShow) Execute(args []string) error {
	setup()

	if len(args) != 1 {
		showParser.WriteHelp(os.Stdout)
		fmt.Println()
		return fmt.Errorf("show: missing required option <tunnelname>\n")
	}

	cert := certificate()
	cl := NewClient(serverAddr, cert)
	tun := cl.Get(args[0])

	if c.Raw {
		log.Println(tun)
	} else {
		cfg, err := configuration.LoadString(tun)
		if err != nil {
			log.Fatal(err)
		}

		if globalOpts.Remap {
			cfg.Remap()
		}

		for _, host := range cfg.Hosts {
			log.Printf("Host %q", host.Name)
			log.Printf("  %s@%s:%d", host.User, host.Addr, host.Port)
			if host.Pass != "" {
				log.Println("  Password authentication")
			}
			if host.Key != "" {
				log.Println("  Key authentication")
			}
			log.Println()
		}
		for _, fwd := range cfg.Forwards {
			log.Printf("Forward %q", fwd.Name)
			for _, line := range fwd.Lines {
				log.Println("  ", line)
			}
			log.Println()
		}
	}

	return nil
}
