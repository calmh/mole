package main

import (
	"bytes"
	"code.google.com/p/go.net/proxy"
	"flag"
	"fmt"
	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/conf"
	"io/ioutil"
	"os"
	"time"
)

func init() {
	addCommand(command{name: "dig", fn: commandDig, descr: msgDigShort, aliases: []string{"connect"}})
}

const keepaliveInterval = 45 * time.Second

func commandDig(args []string) {
	fs := flag.NewFlagSet("dig", flag.ExitOnError)
	local := fs.Bool("l", false, "Local file, not remote tunnel definition")
	qualify := fs.Bool("q", false, "Use <host>.<tunnel> for host aliases instead of just <host>")
	fs.Usage = usageFor(fs, msgDigUsage)
	fs.Parse(args)
	args = fs.Args()

	if len(args) != 1 {
		fs.Usage()
		os.Exit(3)
	}

	// Fail early in case we don't have root since it's always required on
	// platforms where it matters
	requireRoot("dig")

	cfg := loadTunnel(args[0], *local)

	for _, cmt := range cfg.Comments {
		infoln(ansi.Cyan("; " + cmt))
	}
	for _, cmt := range cfg.General.Comments {
		infoln(ansi.Cyan("; " + cmt))
	}

	var addrs []string
	if remapIntfs {
		cfg.Remap()
	} else {
		addrs = missingAddresses(cfg)
		if len(addrs) > 0 {
			addAddresses(addrs)
		}
	}

	infoln(sshPathStr(cfg.General.Main, cfg), "...")

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
		sshConn, err := sshHost(mh, cfg)
		fatalErr(err)
		dialer = sshConn

		go keepalive(cfg, dialer)
	}

	fwdChan := startForwarder(dialer)
	sendForwards(fwdChan, cfg)

	setupHostsFile(args[0], cfg, *qualify)

	shell(fwdChan, cfg, dialer)

	restoreHostsFile(args[0], *qualify)

	if vpn != nil {
		vpn.Stop()
	}

	if !remapIntfs {
		addrs = extraneousAddresses(cfg)
		if len(addrs) > 0 {
			removeAddresses(addrs)
		}
	}

	okln("Done")
	printTotalStats()
}

func keepalive(cfg *conf.Config, dialer Dialer) {
	// Periodically connect to a forward to provide a primitive keepalive mechanism.
	i := 0
	for {
		for _, fwd := range cfg.Forwards {
			for _, line := range fwd.Lines {
				time.Sleep(keepaliveInterval)
				go func(line conf.ForwardLine) {
					x := 0
					if line.Repeat > 0 {
						x = i % line.Repeat
					}
					debugln("keepalive dial", line.DstString(x))
					conn, err := dialer.Dial("tcp", line.DstString(x))
					if err != nil {
						debugln("keepalive dial", err)
					}
					if conn != nil {
						conn.Close()
					}
				}(line)
			}
		}
		i++
	}
}

func loadTunnel(name string, local bool) *conf.Config {
	var err error
	var tun string

	cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))

	if local {
		fd, err := os.Open(name)
		fatalErr(err)
		bs, err := ioutil.ReadAll(fd)
		fatalErr(err)
		tuni, err := authenticated(cl, func() (interface{}, error) { return cl.Deobfuscate(string(bs)) })
		fatalErr(err)
		tun = tuni.(string)
	} else {
		tuni, err := authenticated(cl, func() (interface{}, error) { return cl.Get(name) })
		fatalErr(err)
		tun = tuni.(string)
		tun, err = cl.Deobfuscate(tun)
		fatalErr(err)
	}

	cfg, err := conf.Load(bytes.NewBufferString(tun))
	fatalErr(err)
	return cfg
}

func sendForwards(fwdChan chan<- conf.ForwardLine, cfg *conf.Config) {
	for _, fwd := range cfg.Forwards {
		infoln(ansi.Bold(ansi.Cyan(fwd.Name)))
		for _, cmt := range fwd.Comments {
			infoln(ansi.Cyan("  ; " + cmt))
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
