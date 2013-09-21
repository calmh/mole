package main

import (
	"bytes"
	"code.google.com/p/go.net/proxy"
	"fmt"
	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/conf"
	"github.com/jessevdk/go-flags"
	"os"
	"strings"
)

type cmdDig struct {
	Local        bool `short:"l" long:"local" description:"Local file, not remote tunnel definition"`
	QualifyHosts bool `short:"q" long:"qualify-hosts" description:"Use <host>.<tunnel> for host aliases instead of just <host>"`
}

var digParser *flags.Parser

func init() {
	cmd := cmdDig{}
	digParser = globalParser.AddCommand("dig", msgDigShort, msgDigLong, &cmd)
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

	// Fail early in case we don't have root since it's always required on
	// platforms where it matters
	requireRoot("dig")

	var cfg *conf.Config
	var err error

	if c.Local {
		fd, err := os.Open(args[0])
		fatalErr(err)
		cfg, err = conf.Load(fd)
		fatalErr(err)
	} else {
		cl := NewClient(serverIni.address, serverIni.fingerprint)
		_, err := authenticate(cl)
		fatalErr(err)
		tun, err := cl.Get(args[0])
		fatalErr(err)
		tun, err = cl.Deobfuscate(tun)
		fatalErr(err)
		cfg, err = conf.Load(bytes.NewBufferString(tun))
		fatalErr(err)
	}

	if cfg == nil {
		return fmt.Errorf("no tunnel loaded")
	}

	var addrs []string
	if globalOpts.Remap {
		cfg.Remap()
	} else {
		addrs = missingAddresses(cfg)
		if len(addrs) > 0 {
			addAddresses(addrs)
		}
	}

	infoln(sshPathStr(cfg.General.Main, cfg), "...")

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

	fwdChan := startForwarder(dialer)
	sendForwards(fwdChan, cfg)

	setupHostsFile(args[0], cfg, c.QualifyHosts)

	shell(fwdChan, cfg, dialer)

	restoreHostsFile(args[0], c.QualifyHosts)

	if vpn != nil {
		vpn.Stop()
	}

	if !globalOpts.Remap {
		addrs = extraneousAddresses(cfg)
		if len(addrs) > 0 {
			removeAddresses(addrs)
		}
	}

	okln("Done")
	return nil
}

func sendForwards(fwdChan chan<- conf.ForwardLine, cfg *conf.Config) {
	for _, fwd := range cfg.Forwards {
		infoln(ansi.Underline(fwd.Name))
		if fwd.Comment != "" {
			lines := strings.Split(fwd.Comment, "\\n") // Yes, literal backslash-n
			for i := range lines {
				infoln(ansi.Cyan("  # " + lines[i]))
			}
		}
		for _, line := range fwd.Lines {
			infoln("  " + line.String())
			fwdChan <- line
		}
	}
}

func sshPathStr(hostname string, cfg *conf.Config) string {
	var this string
	if hostID, ok := cfg.HostsMap[hostname]; ok {
		host := cfg.Hosts[hostID]
		this = fmt.Sprintf("ssh://%s@%s", host.User, host.Name)

		if host.Via != "" {
			this = sshPathStr(host.Via, cfg) + " -> " + this
		}

		if host.SOCKS != "" {
			this = "SOCKS://" + host.SOCKS + " -> " + this
		}
	}

	if hostname == cfg.General.Main || hostname == "" {
		if cfg.Vpnc != nil {
			vpnc := fmt.Sprintf("vpnc://%s", cfg.Vpnc["IPSec_gateway"])
			if this == "" {
				return vpnc
			} else {
				this = vpnc + " -> " + this
			}
		} else if cfg.OpenConnect != nil {
			opnc := fmt.Sprintf("openconnect://%s", cfg.OpenConnect["server"])
			if this == "" {
				return opnc
			} else {
				this = opnc + " -> " + this
			}
		}
	}

	return this
}
