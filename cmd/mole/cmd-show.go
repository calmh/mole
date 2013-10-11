package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/calmh/mole/conf"
	"os"
	"strings"
)

func init() {
	commands["show"] = command{showCommand, msgShowShort}
}

func showCommand(args []string) {
	fs := flag.NewFlagSet("show", flag.ExitOnError)
	raw := fs.Bool("r", false, "Show raw tunnel file")
	fs.Usage = usageFor(fs, msgShowUsage)
	fs.Parse(args)
	args = fs.Args()

	if len(args) != 1 {
		fs.Usage()
		os.Exit(3)
	}

	cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))
	res, err := authenticated(cl, func() (interface{}, error) { return cl.Get(args[0]) })
	fatalErr(err)
	tun := res.(string)

	if *raw {
		// No log function, since it must be possible to pipe to a valid file
		fmt.Printf(tun)
	} else {
		cfg, err := conf.Load(bytes.NewBufferString(tun))
		fatalErr(err)

		if remapIntfs {
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
}
