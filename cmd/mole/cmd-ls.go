package main

import (
	"fmt"
	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/table"
	"github.com/jessevdk/go-flags"
	"os"
	"path"
	"regexp"
	"strings"
)

type lsCommand struct {
	Long  bool `short:"l" description:"Long listing"`
	Short bool `short:"s" description:"Short listing (name only)"`
}

var lsParser *flags.Parser

func init() {
	cmd := lsCommand{}
	lsParser = globalParser.AddCommand("ls", msgLsShort, msgLsLong, &cmd)
}

func (c *lsCommand) Usage() string {
	return "[regexp]"
}

func (c *lsCommand) Execute(args []string) error {
	setup()

	var re *regexp.Regexp
	var err error
	if len(args) == 1 {
		re, err = regexp.Compile("(?i)" + args[0])
		fatalErr(err)
	}

	cl := NewClient(serverIni.address, serverIni.fingerprint)
	res, err := authenticated(cl, func() (interface{}, error) { return cl.List() })
	fatalErr(err)
	l := res.([]ListItem)

	var rows [][]string
	header := []string{"TUNNEL", "DESCRIPTION", "HOSTS"}
	if c.Long {
		header = append(header, "VER")
	}
	rows = append(rows, header)

	tunnelCache, _ := os.Create(path.Join(globalOpts.Home, "tunnels.cache"))
	var matched int
	for _, i := range l {
		hosts := strings.Join(i.Hosts, ", ")
		if re == nil || re.MatchString(i.Name) || re.MatchString(i.Description) || re.MatchString(hosts) {
			if tunnelCache != nil {
				fmt.Fprintln(tunnelCache, i.Name)
			}
			if c.Short {
				fmt.Println(i.Name)
			} else {
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

				row := []string{ansi.Bold(ansi.Blue(i.Name)), descr, hosts}
				if c.Long {
					row = append(row, fmt.Sprintf("%d", i.IntVersion))
				}
				rows = append(rows, row)
			}
		}
	}
	if tunnelCache != nil {
		_ = tunnelCache.Close()
	}

	if !c.Short {
		// Never prefix table with log stuff
		fmt.Printf(table.Fmt("lllr", rows))

		if matched != len(l) {
			fmt.Printf(ansi.Faint(" - Matched %d out of %d records\n"), matched, len(l))
		}
	}
	return nil
}
