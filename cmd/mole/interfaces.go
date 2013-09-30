package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/calmh/mole/conf"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var errNoLoopbackFound = errors.New("no loopback interface found")
var keepAddressRe = regexp.MustCompile(`^127\.0\.0\.([0-9]|[0-2][0-9]|3[0-1])$`)

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
	ifconfigAddresses("remove", addrs)
}

func ifconfigAddresses(command string, addrs []string) {
	requireRoot("ifconfig")

	lo := loInterface()
	var cmd bytes.Buffer
	for i := range addrs {
		_, _ = cmd.WriteString(fmt.Sprintf("ifconfig %s %s %s;", lo, command, addrs[i]))
	}

	debugln(cmd.String())
	ifconfig := exec.Command("sh", "-c", cmd.String())
	ifconfig.Stderr = os.Stderr
	ifconfig.Stdout = os.Stdout
	ifconfig.Stdin = os.Stdin
	err := ifconfig.Run()
	fatalErr(err)
}
