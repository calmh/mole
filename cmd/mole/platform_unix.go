// +build darwin linux

package main

import (
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/calmh/mole/conf"
	"github.com/calmh/mole/hosts"
)

var hasRoot bool
var realUid int

func init() {
	hasRoot = syscall.Geteuid() == 0

	sudoUid, ok := syscall.Getenv("SUDO_UID")
	if ok {
		var err error
		realUid, err = strconv.Atoi(sudoUid)
		if err != nil {
			fatalln("SUDO_UID", err)
		}
	} else {
		realUid = syscall.Getuid()
	}

	dropRoot()
}

func requireRoot(reason string) {
	if !hasRoot {
		fatalf(msgErrGainRoot, reason)
	}
}

func becomeRoot() {
	if hasRoot && syscall.Geteuid() != 0 {
		e := syscall.Setreuid(-1, 0)
		fatalErr(e)
	}
}

func dropRoot() {
	if hasRoot && syscall.Geteuid() != realUid {
		e := syscall.Setreuid(-1, realUid)
		fatalErr(e)
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
	becomeRoot()
	var err error
	if qualify {
		err = hosts.ReplaceTagged("mole."+tun, nil)
	} else {
		err = hosts.ReplaceTagged("mole", nil)
	}
	if err != nil {
		warnln(err)
	}
	dropRoot()
}

func addToHostsFile(tag string, domain string, cfg *conf.Config) {
	var entries []hosts.Entry
	for _, fwd := range cfg.Forwards {
		ps := strings.SplitN(fwd.Name, " ", 2)
		name := strings.ToLower(ps[0])
		if domain != "" {
			name = name + "." + domain
		}
		ip := fwd.Lines[0].SrcIP
		entries = append(entries, hosts.Entry{IP: ip, Names: []string{name}})
	}

	requireRoot("update /etc/hosts")
	becomeRoot()
	err := hosts.ReplaceTagged(tag, entries)
	dropRoot()
	fatalErr(err)
}
