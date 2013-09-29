package main

import (
	"flag"
	"fmt"
	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/table"
	"os"
	"path"
	"regexp"
	"strings"
)

func init() {
	commands["ls"] = command{commandLs, msgLsShort}
}

func commandLs(args []string) error {
	fs := flag.NewFlagSet("ls", flag.ExitOnError)
	short := fs.Bool("s", false, "Short listing")
	long := fs.Bool("l", false, "Long listing")
	fs.Usage = usageFor(fs, msgLsUsage)
	fs.Parse(args)
	args = fs.Args()

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
	if *long {
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
			if *short {
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
				} else if len(hosts) > 20 {
					hosts = hosts[:17] + ansi.Faint("...")
				}

				row := []string{ansi.Bold(ansi.Cyan(i.Name)), descr, hosts}
				if *long {
					row = append(row, fmt.Sprintf("%d", i.IntVersion))
				}
				rows = append(rows, row)
			}
		}
	}
	if tunnelCache != nil {
		_ = tunnelCache.Close()
	}

	if !*short {
		// Never prefix table with log stuff
		fmt.Printf(table.Fmt("lllr", rows))

		if matched != len(l) {
			fmt.Printf(ansi.Faint(" - Matched %d out of %d records\n"), matched, len(l))
		}
	}
	return nil
}
