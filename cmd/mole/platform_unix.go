// +build darwin linux

package main

import (
	"os"
	"strings"
	"syscall"

	"github.com/calmh/mole/conf"
	"github.com/calmh/mole/hosts"
)

func requireRoot(reason string) {
	if syscall.Geteuid() != 0 {
		fatalf(msgErrGainRoot, reason)
	}
}

func getHomeDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		fatalln(msgErrNoHome)
	}
	return home
}

func setupHostsFile(tun string, cfg *conf.Config, qualify bool) {
	if qualify {
		addToHostsFile("mole."+tun, tun, cfg)
	} else {
		addToHostsFile("mole", "", cfg)
	}
}

func restoreHostsFile(tun string, qualify bool) {
	var err error
	if qualify {
		err = hosts.ReplaceTagged("mole."+tun, nil)
	} else {
		err = hosts.ReplaceTagged("mole", nil)
	}
	if err != nil {
		warnln(err)
	}
}

func addToHostsFile(tag string, domain string, cfg *conf.Config) {
	var entries []hosts.Entry
	for _, fwd := range cfg.Forwards {
		ps := strings.SplitN(fwd.Name, " ", 2)
		name := strings.ToLower(ps[0])
		if domain != "" {
			name = name + "." + domain
		}
		ip := fwd.Lines[0].Src.Addr.String()
		entries = append(entries, hosts.Entry{IP: ip, Names: []string{name}})
	}

	requireRoot("update /etc/hosts")
	err := hosts.ReplaceTagged(tag, entries)
	fatalErr(err)
}
