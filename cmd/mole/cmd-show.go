package main

import (
	"bytes"
	"fmt"
	"github.com/calmh/mole/conf"
	"github.com/jessevdk/go-flags"
	"os"
	"strings"
)

type showCommand struct {
	Raw bool `short:"r" long:"raw" description:"Show raw, not parsed, tunnel file"`
}

var showParser *flags.Parser

func init() {
	cmd := showCommand{}
	showParser = globalParser.AddCommand("show", msgShowShort, msgShowLong, &cmd)
}

func (c *showCommand) Usage() string {
	return "<tunnelname> [show-OPTIONS]"
}

func (c *showCommand) Execute(args []string) error {
	setup()

	if len(args) != 1 {
		showParser.WriteHelp(os.Stdout)
		infoln()
		fatalln("show: missing required option <tunnelname>")
	}

	cl := NewClient(serverIni.address, serverIni.fingerprint)
	res, err := authenticated(cl, func() (interface{}, error) { return cl.Get(args[0]) })
	fatalErr(err)
	tun := res.(string)

	if c.Raw {
		// No log function, since it must be possible to pipe to a valid file
		fmt.Printf(tun)
	} else {
		cfg, err := conf.Load(bytes.NewBufferString(tun))
		fatalErr(err)

		if globalOpts.Remap {
			cfg.Remap()
		}

		for _, host := range cfg.Hosts {
			infof("Host %q", host.Name)
			infof("  %s@%s:%d", host.User, host.Addr, host.Port)
			if host.Pass != "" {
				infoln("  Password authentication")
			}
			if host.Key != "" {
				infoln("  Key authentication")
			}
		}
		for _, fwd := range cfg.Forwards {
			infof("Forward %q", fwd.Name)
			if fwd.Comment != "" {
				lines := strings.Split(fwd.Comment, "\\n") // Yes, literal backslash-n
				for i := range lines {
					infoln("  # " + lines[i])
				}
			}
			for _, line := range fwd.Lines {
				infoln("  " + line.String())
			}
		}
	}

	return nil
}
