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
	"strings"
	"time"
)

func init() {
	commands["dig"] = command{commandDig, msgDigShort}
}

func commandDig(args []string) error {
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

	var err error

	cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))
	_, err = authenticated(cl, func() (interface{}, error) { return cl.Ping() })
	fatalErr(err)

	var tun string
	if *local {
		fd, err := os.Open(args[0])
		fatalErr(err)
		bs, err := ioutil.ReadAll(fd)
		fatalErr(err)
		tun, err = cl.Deobfuscate(string(bs))
		fatalErr(err)
	} else {
		tun, err = cl.Get(args[0])
		fatalErr(err)
	}

	tun, err = cl.Deobfuscate(tun)
	fatalErr(err)
	cfg, err := conf.Load(bytes.NewBufferString(tun))
	fatalErr(err)

	if cfg == nil {
		return fmt.Errorf("no tunnel loaded")
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

		// A REALY dirty keepalive mechanism.
		sess, err := sshConn.NewSession()
		fatalErr(err)
		stdout, err := sess.StdoutPipe()
		fatalErr(err)
		stderr, err := sess.StderrPipe()
		fatalErr(err)
		stdin, err := sess.StdinPipe()
		fatalErr(err)
		err = sess.Shell()
		fatalErr(err)

		go func() {
			bs := make([]byte, 1024)
			for {
				debugln("shell read stdout")
				_, err := stdout.Read(bs)
				fatalErr(err)
			}
		}()

		go func() {
			bs := make([]byte, 1024)
			for {
				debugln("shell read stderr")
				_, err := stderr.Read(bs)
				fatalErr(err)
			}
		}()

		go func() {
			for {
				time.Sleep(30 * time.Second)
				debugln("shell write")
				_, err := stdin.Write([]byte("\n"))
				fatalErr(err)
			}
		}()
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

	return nil
}

func sendForwards(fwdChan chan<- conf.ForwardLine, cfg *conf.Config) {
	for _, fwd := range cfg.Forwards {
		infoln(ansi.Bold(ansi.Cyan(fwd.Name)))
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
