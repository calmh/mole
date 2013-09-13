package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"nym.se/mole/conf"
)

var errNoLoopbackFound = errors.New("no loopback interface found")
var keepAddressRe = regexp.MustCompile(`^127\.0\.0\.([0-9]|[0-2][0-9]|3[0-1])$`)

func loInterface() string {
	intfs, e := net.Interfaces()
	if e != nil {
		log.Fatal(e)
	}

	for _, intf := range intfs {
		if intf.Flags&net.FlagLoopback == net.FlagLoopback {
			if globalOpts.Debug {
				log.Printf("loopback interface on %q", intf.Name)
			}
			return intf.Name
		}
	}

	log.Fatal(errNoLoopbackFound)
	return "" // Unreachable
}

func currentAddresses() []string {
	addrs, e := net.InterfaceAddrs()
	if e != nil {
		log.Fatal(e)
	}

	cur := make([]string, len(addrs))
	for i := range addrs {
		s := addrs[i].String()
		ps := strings.SplitN(s, "/", 2)
		cur[i] = ps[0]
	}

	debug("current interface addresses: %v", cur)
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

	if globalOpts.Debug {
		log.Printf("missing local addresses: %v", missing)
	}
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

	if globalOpts.Debug {
		log.Printf("extraneous interface addresses: %v", extra)
	}

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
		cmd.WriteString(fmt.Sprintf("ifconfig %s %s %s;", lo, command, addrs[i]))
	}

	debug(cmd.String())
	ifconfig := exec.Command("sh", "-c", cmd.String())
	ifconfig.Stderr = os.Stderr
	ifconfig.Stdout = os.Stdout
	ifconfig.Stdin = os.Stdin
	e := ifconfig.Run()
	if e != nil {
		log.Fatal(e)
	}
}
