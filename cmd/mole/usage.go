package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/calmh/mole/ansi"
)

func optionTable(w io.Writer, rows [][]string) {
	tw := tabwriter.NewWriter(w, 2, 4, 2, ' ', 0)
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				tw.Write([]byte("\t"))
			}
			tw.Write([]byte(cell))
		}
		tw.Write([]byte("\n"))
	}
	tw.Flush()
}

func usageFor(fs *flag.FlagSet, usage string) func() {
	return func() {
		var b bytes.Buffer
		b.WriteString(ansi.Bold("Usage:") + "\n  " + usage + "\n")

		var options [][]string
		fs.VisitAll(func(f *flag.Flag) {
			var opt = "  -" + f.Name

			if f.DefValue != "false" {
				opt += "=" + f.DefValue
			}

			options = append(options, []string{opt, f.Usage})
		})

		if len(options) > 0 {
			b.WriteString("\n" + ansi.Bold("Options:") + "\n")
			optionTable(&b, options)
		}

		fmt.Println(b.String())

	}
}

func mainUsage(w io.Writer) {
	tw := tabwriter.NewWriter(w, 2, 4, 2, ' ', 0)

	fmt.Fprintln(w, ansi.Bold("Commands:"))
	var cmds []string
	for _, cmd := range commandList {
		cmds = append(cmds, cmd.name)
	}
	sort.Strings(cmds)
	for _, name := range cmds {
		cmd := commandMap[name]
		if sn := cmd.descr; sn != "" {
			alias := ""
			// Ignore undocumented commands
			if len(cmd.aliases) > 0 {
				alias = " (" + strings.Join(cmd.aliases, ", ") + ")"
			}
			tw.Write([]byte(fmt.Sprintf("  %s%s\t%s\n", ansi.Bold(ansi.Cyan(name)), ansi.Cyan(alias), sn)))
		}
	}
	tw.Flush()
	fmt.Fprintf(w, "\n  Commands can be abbreviated to their unique prefix.\n\n")

	fmt.Fprintln(w, ansi.Bold("Examples:"))
	examples := [][]string{
		{"  mole ls", "# show all available tunnels"},
		{"  mole l foo", "# show all available tunnels matching the regexp \"foo\""},
		{"  mole show foo", "# show the hosts and forwards in the tunnel \"foo\""},
		{"  sudo mole dig foo", "# dig the tunnel \"foo\""},
		{"  sudo mole -d d foo", "# dig the tunnel \"foo\", while showing debug output"},
		{"  mole push foo.ini", "# create or update the \"foo\" tunnel from a local file"},
		{"  mole install", "# list packages available for installation"},
		{"  mole ins vpnc", "# install a package named vpnc"},
		{"  mole up -force", "# perform a forced up/downgrade to the server version"},
	}
	optionTable(w, examples)
	w.Write([]byte("\n"))
}
