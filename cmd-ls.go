package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
)

type cmdls struct{}

var lsParser *flags.Parser

func init() {
	cmd := cmdls{}
	lsParser = globalParser.AddCommand("ls", "ls a tunnel", "'ls' connects to a remote destination and sets up configured local TCP tunnels", &cmd)
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
		if err != nil {
			log.Fatal(err)
		}
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
				descr = descr + magenta(" (vpnc)")
			} else if i.OpenConnect {
				descr = descr + green(" (opnc)")
			} else if i.Socks {
				descr = descr + yellow(" (socks)")
			}

			if hosts == "" {
				hosts = faint("(local forward)")
			}

			rows = append(rows, []string{bold(blue(i.Name)), descr, hosts})
		}
	}

	// Never prefix table with log stuff
	fmt.Printf(table(rows))

	if matched != len(l) {
		fmt.Printf(faint(" - Matched %d out of %d records\n"), matched, len(l))
	}

	return nil
}
