package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"code.google.com/p/go.net/proxy"
	"github.com/jessevdk/go-flags"
	"nym.se/mole/ansi"
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
		cl := NewClient(serverIni.address, cert)
		tun := cl.Deobfuscate(cl.Get(args[0]))
		cfg, err = conf.LoadString(tun)
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

	var vpn VPN

	if cfg.Vpnc != nil {
		vpn = startVpn("vpnc", cfg)
	} else if cfg.OpenConnect != nil {
		vpn = startVpn("openconnect", cfg)
	}

	var dialer Dialer = proxy.Direct
	if mh := cfg.General.Main; mh != "" {
		dialer = sshHost(mh, cfg)
	}

	fwdChan := startForwarder(dialer)
	sendForwards(fwdChan, cfg)

	shell(fwdChan, cfg, dialer)

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

func sshHost(host string, cfg *conf.Config) Dialer {
	h := cfg.Hosts[cfg.HostsMap[host]]
	var conn net.Conn
	var err error
	if h.Via != "" {
		conn, err = sshHost(h.Via, cfg).Dial("tcp", fmt.Sprintf("%s:%d", h.Addr, h.Port))
	} else {
		var dialer Dialer = proxy.Direct
		if h.SOCKS != "" {
			dialer, err = proxy.SOCKS5("tcp", h.SOCKS, nil, proxy.Direct)
			fatalErr(err)
		}
		conn, err = dialer.Dial("tcp", fmt.Sprintf("%s:%d", h.Addr, h.Port))
	}
	fatalErr(err)
	return sshOnConn(conn, h)
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
