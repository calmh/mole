package table

import (
	"bytes"

	"nym.se/mole/ansi"
)

const (
	paddingChars       = "                                                                          "
	intraColumnPadding = 2
)

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

	var buf bytes.Buffer
	for r, row := range rows {
		for c, cell := range row {
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
