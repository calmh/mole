package main

import (
	"code.google.com/p/go.net/proxy"
	"flag"
	"fmt"
	"os"
)

func init() {
	commands["test"] = command{testCommand, msgTestShort}
}

func testCommand(args []string) {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	local := fs.Bool("l", false, "Local file, not remote tunnel definition")
	fs.Usage = usageFor(fs, msgTestUsage)
	fs.Parse(args)
	args = fs.Args()

	if len(args) != 1 {
		fs.Usage()
		os.Exit(3)
	}

	// Fail early in case we don't have root since it's always required on
	// platforms where it matters
	requireRoot("test")

	cfg := loadTunnel(args[0], *local)

	var vpn VPN
	var err error

	if cfg.Vpnc != nil {
		vpn, err = startVpn("vpnc", cfg)
		fatalErr(err)
	} else if cfg.OpenConnect != nil {
		vpn, err = startVpn("openconnect", cfg)
		fatalErr(err)
	}

	var dialer Dialer = proxy.Direct
	if mh := cfg.General.Main; mh != "" {
		dialer, err = sshHost(mh, cfg)
		fatalErr(err)
	}

	var ok, failed int
	results := testForwards(dialer, cfg)
	for result := range results {
		for _, forwardres := range result.results {
			if forwardres.err == nil {
				ok++
			} else {
				failed++
			}
		}
	}

	if vpn != nil {
		vpn.Stop()
	}

	msg := fmt.Sprintf("%d of %d port forwards connect successfully", ok, ok+failed)
	if ok > failed {
		okln(msg)
	} else {
		warnln(msg)
		os.Exit(1)
	}
}
