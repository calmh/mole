// Package table formats an ASCII/ANSI table with dynamic column widths.
package table

import (
	"bytes"
	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/termsize"
)

const (
	paddingChars       = "                                                                          "
	intraColumnPadding = 2
)

// Fmt formats the given rows into a pretty table and returns the string ready
// to be printed. The fmt parameter is a string of "l" and "r" characters
// indicating the requested alignment of each column. Table cells can contain
// ANSI formatting. The header row will be underlined. Example:
//
//   rows := [][]string{{"NAME", "VALUE"}, {"Test", "4.5"}, {"Other", "13.2"}}
//   tab := table.Fmt("lr", rows)
//   fmt.Println(tab)
//
func Fmt(fmt string, rows [][]string) string {
	cols := len(rows[0])
	width := make([]int, cols)
	for _, row := range rows {
		for i, cell := range row {
			if l := ansi.Strlen(cell); l > width[i] {
				width[i] = l
			}
		}
	}

	tw := termsize.Columns()
	cw := 0
	maxCols := 0
	for _, w := range width {
		cw += w
		if cw > tw {
			break
		}
		maxCols++
	}

	var buf bytes.Buffer
	for r, row := range rows {
		for c, cell := range row {
			if c >= maxCols {
				break
			}
			if r == 0 {
				cell = ansi.Underline(cell)
			}

			l := ansi.Strlen(cell)
			pad := paddingChars[:width[c]-l]
			if r == 0 {
				pad = ansi.Underline(pad)
			}

			if r == 0 || fmt[c] == 'l' {
				buf.WriteString(cell)
				buf.WriteString(pad)
			} else if fmt[c] == 'r' {
				buf.WriteString(pad)
				buf.WriteString(cell)
			}

			if c < cols-1 {
				buf.WriteString(paddingChars[:intraColumnPadding])
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}
