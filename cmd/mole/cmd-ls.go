package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
	"nym.se/mole/ansi"
	"nym.se/mole/table"
)

type cmdls struct{}

var lsParser *flags.Parser

func init() {
	cmd := cmdls{}
	lsParser = globalParser.AddCommand("ls", "List available tunnels", "'ls' lists tunnels while optionally filtering on a provided regular expression", &cmd)
}

func (c *cmdls) Usage() string {
	return "[regexp]"
}

func (c *cmdls) Execute(args []string) error {
	setup()

	var re *regexp.Regexp
	var err error
	if len(args) == 1 {
		re, err = regexp.Compile("(?i)" + args[0])
		fatalErr(err)
	}

	cert := certificate()
	cl := NewClient(serverAddr, cert)
	l := cl.List()

	var rows [][]string
	rows = append(rows, []string{"TUNNEL", "DESCRIPTION", "HOSTS"})

	var matched int
	for _, i := range l {
		hosts := strings.Join(i.Hosts, ", ")
		if re == nil || re.MatchString(i.Name) || re.MatchString(i.Description) || re.MatchString(hosts) {
			matched++

			descr := i.Description
			if i.Vpnc {
				descr = descr + ansi.Magenta(" (vpnc)")
			} else if i.OpenConnect {
				descr = descr + ansi.Green(" (opnc)")
			} else if i.Socks {
				descr = descr + ansi.Yellow(" (socks)")
			}

			if hosts == "" {
				hosts = ansi.Faint("(local forward)")
			}

			rows = append(rows, []string{ansi.Bold(ansi.Blue(i.Name)), descr, hosts})
		}
	}

	// Never prefix table with log stuff
	fmt.Printf(table.Fmt("lll", rows))

	if matched != len(l) {
		fmt.Printf(ansi.Faint(" - Matched %d out of %d records\n"), matched, len(l))
	}

	return nil
}
