package main

import (
	"bytes"
	"flag"
	"fmt"
	"mole/ansi"
	"io"
	"text/tabwriter"
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
