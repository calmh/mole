package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.net/proxy"
	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/conf"
	"github.com/calmh/mole/upgrade"
)

func init() {
	addCommand(command{name: "dig", fn: commandDig, descr: msgDigShort, aliases: []string{"connect"}})
}

var keepaliveInterval = 120 * time.Second

func commandDig(args []string) {
	fs := flag.NewFlagSet("dig", flag.ExitOnError)
	local := fs.Bool("l", false, "Local file, not remote tunnel definition")
	qualify := fs.Bool("q", false, "Use <host>.<tunnel> for host aliases instead of just <host>")
	noVerify := fs.Bool("n", false, "Don't verify connectivity")
	fs.DurationVar(&keepaliveInterval, "keepalive", keepaliveInterval, "SSH server alive timeout")
	fs.Usage = usageFor(fs, msgDigUsage)
	fs.Parse(args)
	args = fs.Args()

	if l := len(args); l < 1 || l > 2 {
		fs.Usage()
		exit(3)
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
		atExit(func() {
			addrs := extraneousAddresses(cfg)
			if len(addrs) > 0 {
				removeAddresses(addrs)
			}
		})
	}

	if mh := fs.Arg(1); mh != "" {
		_, ok := cfg.HostsMap[mh]
		if !ok {
			fatalf(msgDigNoHost, mh)
		}
		cfg.General.Main = mh
		warnln(msgDigWarnMainHost)
		*noVerify = true
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
	atExit(func() {
		if vpn != nil {
			vpn.Stop()
		}
	})

	var dialer Dialer = proxy.Direct
	if mh := cfg.General.Main; mh != "" {
		sshConn, err := sshHost(mh, cfg)
		fatalErr(err)
		dialer = sshConn

		startKeepalive(dialer)
	}

	fwdChan := startForwarder(dialer)
	sendForwards(fwdChan, cfg)

	setupHostsFile(args[0], cfg, *qualify)
	atExit(func() {
		restoreHostsFile(args[0], *qualify)
	})

	if !*noVerify {
		go verify(dialer, cfg)
	}

	go autoUpgrade()

	shell(fwdChan, cfg, dialer)

	okln("Done")
	printTotalStats()
}

func startKeepalive(dialer Dialer) {
	conn := dialer.(*ssh.ClientConn)
	kc := make(chan time.Duration)

	go func() {
		for {
			t0 := time.Now()
			err := conn.CheckServerAlive()
			fatalErr(err)
			kc <- time.Since(t0)

			time.Sleep(keepaliveInterval)
		}
	}()

	go func() {
		for {
			select {
			case t := <-kc:
				debugf("keepalive response in %.01f ms", t.Seconds()*1000)
			case <-time.After(2*keepaliveInterval + 2*time.Second):
				fatalln(msgKeepaliveTimeout)
			}
		}
	}()
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

func verify(dialer Dialer, cfg *conf.Config) {
	okln(msgTesting)
	minRtt := float64(1e100)
	allFwd, okFwd := 0, 0
	results := testForwards(dialer, cfg)
	for res := range results {
		for _, line := range res.results {
			if line.err == nil {
				if line.ms < minRtt {
					minRtt = line.ms
				}
				okFwd++
			}
			allFwd++
		}
	}
	if okFwd == 0 {
		fatalf(msgTunnelVerifyFailed, allFwd)
	} else if float64(okFwd)/float64(allFwd) < 0.5 || minRtt > 250 {
		warnf(msgTunnelRtt, minRtt, okFwd, allFwd)
	} else {
		okf(msgTunnelRtt, minRtt, okFwd, allFwd)
	}
}

func autoUpgrade() {
	if moleIni.Get("upgrades", "automatic") == "no" {
		debugln("automatic upgrades disabled")
		return
	}

	// Only do the actual upgrade once we've been running for a while
	time.Sleep(10 * time.Second)
	build, err := latestBuild()
	if err == nil {
		bd := time.Unix(int64(build.BuildStamp), 0)
		if isNewer := bd.Sub(buildDate).Seconds() > 0; isNewer {
			err = upgrade.UpgradeTo(build)
			if err == nil {
				if moleIni.Get("upgrades", "automatic") != "yes" {
					infoln(msgAutoUpgrades)
				}
				okf(msgUpgraded, build.Version)
			} else {
				warnln("Automatic upgrade failed:", err)
			}
		}
	}
}
