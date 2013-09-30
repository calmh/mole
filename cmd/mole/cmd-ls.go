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
	header := []string{"TUNNEL", "FLAGS", "DESCRIPTION"}
	if *long {
		header = append(header, "HOSTS", "VER")
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

				flags := ""
				if i.Vpnc {
					flags += "v"
				} else {
					flags += "·"
				}
				if i.OpenConnect {
					flags += "o"
				} else {
					flags += "·"
				}
				if i.Socks {
					flags += "s"
				} else {
					flags += "·"
				}

				if hosts == "" {
					hosts = "-"
					flags += "l"
				} else {
					flags += "·"
				}

				row := []string{i.Name, flags, i.Description}
				if *long {
					row = append(row, hosts, fmt.Sprintf("%.01f", float64(i.IntVersion)/100))
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
		fmt.Printf(table.FmtFunc("llllr", rows, tableFormatter))
		if matched != len(l) {
			fmt.Printf(ansi.Faint(" - Matched %d out of %d records\n"), matched, len(l))
		}
	}
	return nil
}

func tableFormatter(cell string, row, col, flags int) string {
	if row == 0 {
		return ansi.Underline(cell)
	} else if col == 0 {
		return ansi.Bold(ansi.Cyan(cell))
	} else if flags&table.Truncated != 0 {
		return cell[:len(cell)-1] + ansi.Red(">")
	}
	return cell
}
