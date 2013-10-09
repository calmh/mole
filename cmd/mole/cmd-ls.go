package main

import (
	"flag"
	"fmt"
	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/conf"
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
	fs := flag.NewFlagSet("ls", flag.ContinueOnError)
	short := fs.Bool("s", false, "Short listing")
	long := fs.Bool("l", false, "Long listing")
	fs.Usage = usageFor(fs, msgLsUsage)
	err := fs.Parse(args)
	if err != nil {
		fmt.Println(ansi.Bold("Feature Flags:"))
		fmt.Println(msgLsFlags)
		os.Exit(3)
	}
	args = fs.Args()

	var re *regexp.Regexp
	if len(args) == 1 {
		re, err = regexp.Compile("(?i)" + args[0])
		fatalErr(err)
	}

	cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))
	res, err := authenticated(cl, func() (interface{}, error) { return cl.List() })
	fatalErr(err)
	l := res.([]ListItem)

	var rows [][]string
	var header []string
	var format string
	var hasFeatureFlags bool

	for _, i := range l {
		if i.Features != 0 {
			hasFeatureFlags = true
			break
		}
	}

	if hasFeatureFlags {
		header = []string{"TUNNEL", "FLAGS", "DESCRIPTION"}
		format = "lll"
	} else {
		header = []string{"TUNNEL", "DESCRIPTION"}
		format = "ll"
	}
	if *long {
		header = append(header, "HOSTS", "VER")
		format += "lr"
	}

	rows = [][]string{header}

	tunnelCache, _ := os.Create(path.Join(homeDir, "tunnels.cache"))

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

				row := []string{i.Name}

				if hasFeatureFlags {
					flags := ""
					var spacer = "·"
					if i.Features&conf.FeatureError != 0 {
						flags = strings.Repeat(spacer, 4) + "E"
					} else {
						if i.Features&conf.FeatureVpnc != 0 {
							flags += "v"
						} else if i.Features&conf.FeatureOpenConnect != 0 {
							flags += "o"
						} else {
							flags += spacer
						}

						if i.Features&conf.FeatureSshKey != 0 {
							flags += "k"
						} else {
							flags += spacer
						}

						if i.Features&conf.FeatureSshPassword != 0 {
							flags += "p"
						} else {
							flags += spacer
						}

						if i.Features&conf.FeatureSocks != 0 {
							flags += "s"
						} else if i.Features&conf.FeatureLocalOnly != 0 {
							flags += "l"
						} else {
							flags += spacer
						}

						if i.Features & ^(conf.FeatureError|conf.FeatureSshKey|conf.FeatureSshPassword|conf.FeatureLocalOnly|conf.FeatureVpnc|conf.FeatureOpenConnect|conf.FeatureSocks) != 0 {
							flags += "U"
						} else {
							flags += spacer
						}
					}
					row = append(row, flags)
				}

				row = append(row, i.Description)

				if *long {
					ver := fmt.Sprintf("%.01f", float64(i.IntVersion)/100)
					if hosts == "" {
						hosts = "-"
					}
					row = append(row, hosts, ver)
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
		fmt.Printf(table.FmtFunc(format, rows, tableFormatter))
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
