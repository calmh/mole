package main

import (
	"errors"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"mole/conf"
)

var errNoLoopbackFound = errors.New("no loopback interface found")
var keepAddressRe = regexp.MustCompile(`^(127\.0\.0\.([0-9]|[0-2][0-9]|3[0-1])|::1)$`)

func loInterface() string {
	intfs, err := net.Interfaces()
	fatalErr(err)

	for _, intf := range intfs {
		if intf.Flags&net.FlagLoopback == net.FlagLoopback {
			debugf("loopback interface on %q", intf.Name)
			return intf.Name
		}
	}

	fatalln(errNoLoopbackFound)
	return "" // Unreachable
}

func currentAddresses() []string {
	addrs, err := net.InterfaceAddrs()
	fatalErr(err)

	cur := make([]string, len(addrs))
	for i := range addrs {
		s := addrs[i].String()
		ps := strings.SplitN(s, "/", 2)
		cur[i] = ps[0]
	}

	debugf("current interface addresses: %v", cur)
	return cur
}

func missingAddresses(cfg *conf.Config) []string {
	current := currentAddresses()
	wanted := cfg.SourceAddresses()

	curMap := make(map[string]bool)
	for _, ip := range current {
		curMap[ip] = true
	}

	var missing []string
	for _, ip := range wanted {
		if ip[0] == '[' {
			ip = ip[1 : len(ip)-1]
		}
		if !curMap[ip] {
			missing = append(missing, ip)
		}
	}

	debugf("missing local addresses: %v", missing)
	return missing
}

func extraneousAddresses(cfg *conf.Config) []string {
	added := cfg.SourceAddresses()
	addedMap := make(map[string]bool)
	for _, ip := range added {
		if ip[0] == '[' {
			ip = ip[1 : len(ip)-1]
		}
		addedMap[ip] = true
	}

	cur := currentAddresses()
	var extra []string
	for _, ip := range cur {
		if addedMap[ip] && !keepAddressRe.MatchString(ip) {
			extra = append(extra, ip)
		}
	}

	debugf("extraneous interface addresses: %v", extra)
	return extra
}

func addAddresses(addrs []string) {
	ifconfigAddresses("add", addrs)
}

func removeAddresses(addrs []string) {
	if runtime.GOOS == "darwin" {
		ifconfigAddresses("remove", addrs)
	} else {
		ifconfigAddresses("del", addrs)
	}
}

func ifconfigAddresses(command string, addrs []string) {
	requireRoot("ifconfig")

	lo := loInterface()
	for _, addr := range addrs {
		debugln("ifconfig", lo, command, addr)
		out, err := exec.Command("ifconfig", lo, command, addr).CombinedOutput()
		if err != nil {
			os.Stdout.Write(out)
			fatalErr(err)
		}
	}
}
