package main

import (
	"bytes"
	"regexp"
	"unicode/utf8"
)

const (
	paddingChars       = "                                                                          "
	intraColumnPadding = 2
)

var re = regexp.MustCompile("\033.+?m")

func ansiRuneLen(s string) int {
	cleaned := re.ReplaceAllString(s, "")
	return utf8.RuneCountInString(cleaned)
}

func tablef(fmt string, rows [][]string) string {
	cols := len(rows[0])
	width := make([]int, cols)
	for _, row := range rows {
		for i, cell := range row {
			if l := ansiRuneLen(cell); l > width[i] {
				width[i] = l
			}
		}
	}

	var buf bytes.Buffer
	for r, row := range rows {
		for c, cell := range row {
			if r == 0 {
				cell = underline(cell)
			}

			l := ansiRuneLen(cell)
			pad := paddingChars[:width[c]-l]
			if r == 0 {
				pad = underline(pad)
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
