// Package table formats an ASCII/ANSI table with dynamic column widths.
package table

import (
	"bytes"
	"github.com/calmh/mole/termsize"
	"strings"
	"unicode/utf8"
)

const (
	intraColumnPadding = 2
)

const (
	// The cell has been padded
	Padded = 1 << iota

	// The cell has been truncated
	Truncated
)

// A Formatter is called to apply string formatting such as ANSI escape
// sequences just before printing a cell. The parameters are the cell text
// (after padding or truncation), the row and column index (zero based) and a
// set of flags. possible flags are Padded (the cell has been space padded on
// the right or left) and Truncated (the cell has been truncated on the
// right). The returned string must have unchanged (visible) length.
type Formatter func(cell string, row int, col int, flags int) string

func identity(cell string, r, c, f int) string {
	return cell
}

// Fmt formats the given rows into a pretty table and returns the string ready
// to be printed. The fmt parameter is a string of "l" and "r" characters
// indicating the requested alignment of each column. Example:
//
//   rows := [][]string{{"NAME", "VALUE"}, {"Test", "4.5"}, {"Other", "13.2"}}
//   tab := table.Fmt("lr", rows)
//   fmt.Println(tab)
//
func Fmt(fmt string, rows [][]string) string {
	return FmtFunc(fmt, rows, identity)
}

// FmtFunc works as Fmt but calls the formatting function fn on each cell
// before printing.
func FmtFunc(fmt string, rows [][]string, fn Formatter) string {
	cols := len(rows[0])
	width := make([]int, cols)
	for _, row := range rows {
		for i, cell := range row {
			if l := utf8.RuneCountInString(cell); l > width[i] {
				width[i] = l
			}
		}
	}

	termWidth := termsize.Columns()
	totWidth := func() (cw int) {
		for _, w := range width {
			cw += w
		}
		cw += intraColumnPadding * (len(width) - 1)
		return cw
	}

	for {
		tw := totWidth()
		if tw <= termWidth {
			break
		}

		// Make wider columns narrower first
		for i := range width {
			if width[i] >= tw/len(width) {
				width[i]--
			}
		}
	}

	var buf bytes.Buffer
	for r, row := range rows {
		for c, cell := range row {
			if width[c] == 0 {
				continue
			}

			flags := 0

			// UTF8 len
			l := utf8.RuneCountInString(cell)
			var pad string
			if l < width[c] {
				pad = strings.Repeat(" ", width[c]-l)
				flags |= Padded
			} else if l > width[c] {
				// BUG(jb): When there are multibyte UTF-8 characters we cut the string shorter than necessary.
				cell = cell[:width[c]]
				flags |= Truncated
			}

			if r == 0 || fmt[c] == 'l' {
				cell = cell + pad
			} else if fmt[c] == 'r' {
				cell = pad + cell
			}

			cell = fn(cell, r, c, flags)
			buf.WriteString(cell)

			if c < cols-1 {
				buf.WriteString(strings.Repeat(" ", intraColumnPadding))
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}
