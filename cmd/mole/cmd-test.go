package main

import (
	"bytes"
	"code.google.com/p/go.net/proxy"
	"flag"
	"fmt"
	"github.com/calmh/mole/conf"
	"os"
)

func init() {
	commands["test"] = command{testCommand, msgTestShort}
}

func testCommand(args []string) error {
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

	var cfg *conf.Config
	var err error

	if *local {
		fd, err := os.Open(args[0])
		fatalErr(err)
		cfg, err = conf.Load(fd)
		fatalErr(err)
	} else {
		cl := NewClient(serverIni.address, serverIni.fingerprint)
		res, err := authenticated(cl, func() (interface{}, error) { return cl.Get(args[0]) })
		fatalErr(err)
		tun := res.(string)
		fatalErr(err)
		tun, err = cl.Deobfuscate(tun)
		fatalErr(err)
		cfg, err = conf.Load(bytes.NewBufferString(tun))
		fatalErr(err)
	}

	if cfg == nil {
		return fmt.Errorf("no tunnel loaded")
	}

	var vpn VPN

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
	return nil
}
