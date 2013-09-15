package ansi

import (
	"regexp"
	"unicode/utf8"
)

var ansiRe = regexp.MustCompile("\033.+?m")

func Strlen(s string) int {
	cleaned := ansiRe.ReplaceAllString(s, "")
	return utf8.RuneCountInString(cleaned)
}
