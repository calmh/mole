package usage

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/calmh/mole/ansi"
	"io"
	"text/tabwriter"
)

func For(fs *flag.FlagSet, usage string) func() {
	return func() {
		var b bytes.Buffer
		b.WriteString(ansi.Bold("Usage:") + "\n  " + usage + "\n")

		var options [][]string
		visitor := func(f *flag.Flag) {
			var opt = "  -" + f.Name

			if f.DefValue != "false" {
				opt += "=VAL"
			}

			def := fmt.Sprintf("(%q)", f.DefValue)
			options = append(options, []string{opt, f.Usage, def})
		}

		fs.VisitAll(visitor)

		if len(options) > 0 {
			b.WriteString("\n" + ansi.Bold("Options:") + "\n")
			Table(&b, options)
		}

		fmt.Println(b.String())
	}
}

func Table(w io.Writer, rows [][]string) {
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
