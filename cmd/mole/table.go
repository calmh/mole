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

func table(rows [][]string) string {
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
			buf.WriteString(cell)

			l := ansiRuneLen(cell)
			pad := paddingChars[:width[c]-l]
			if r == 0 {
				pad = underline(pad)
			}
			buf.WriteString(pad)

			if c < cols-1 {
				buf.WriteString(paddingChars[:intraColumnPadding])
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}
