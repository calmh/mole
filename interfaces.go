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

	"github.com/calmh/mole/configuration"
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

	if globalOpts.Debug {
		log.Printf("current interface addresses: %v", cur)
	}
	return cur
}

func missingAddresses(cfg *configuration.Config) []string {
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

func extraneousAddresses(cfg *configuration.Config) []string {
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
	lo := loInterface()
	var cmd bytes.Buffer
	for i := range addrs {
		cmd.WriteString(fmt.Sprintf("ifconfig %s %s %s;", lo, command, addrs[i]))
	}

	if globalOpts.Debug {
		log.Println(cmd.String())
	}
	args := []string{"-p", "(sudo) Account password, to invoke \"ifconfig " + command + "\": ", "sh", "-c", cmd.String()}
	sudo := exec.Command("sudo", args...)
	sudo.Stderr = os.Stderr
	sudo.Stdout = os.Stdout
	sudo.Stdin = os.Stdin
	e := sudo.Run()
	if e != nil {
		log.Fatal(e)
	}
}
