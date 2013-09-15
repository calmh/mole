package ansi

import (
	"regexp"
	"unicode/utf8"
)

const (
	ansiBold         = "\033[1m"
	ansiFaint        = "\033[2m"
	ansiBoldOff      = "\033[22m"
	ansiUnderline    = "\033[4m"
	ansiUnderlineOff = "\033[24m"
	ansiFgBlack      = "\033[30m"
	ansiFgRed        = "\033[31m"
	ansiFgGreen      = "\033[32m"
	ansiFgYellow     = "\033[33m"
	ansiFgBlue       = "\033[34m"
	ansiFgMagenta    = "\033[35m"
	ansiFgCyan       = "\033[36m"
	ansiFgReset      = "\033[39m"
	ansiKillLine     = "\033[K"
	ansiHideCursor   = "\033[?25l"
	ansiShowCursor   = "\033[?25h"
)

var (
	disabled bool
	ansiRe   = regexp.MustCompile("\033.+?[mKlh]")
)

func Disable() {
	disabled = true
}

func Strlen(s string) int {
	cleaned := ansiRe.ReplaceAllString(s, "")
	return utf8.RuneCountInString(cleaned)
}

func Bold(s string) string {
	if disabled {
		return s
	}
	return ansiBold + s + ansiBoldOff
}

func Faint(s string) string {
	if disabled {
		return s
	}
	return ansiFaint + s + ansiBoldOff
}

func Black(s string) string {
	if disabled {
		return s
	}
	return ansiFgBlack + s + ansiFgReset
}

func Red(s string) string {
	if disabled {
		return s
	}
	return ansiFgRed + s + ansiFgReset
}

func Green(s string) string {
	if disabled {
		return s
	}
	return ansiFgGreen + s + ansiFgReset
}

func Yellow(s string) string {
	if disabled {
		return s
	}
	return ansiFgYellow + s + ansiFgReset
}

func Blue(s string) string {
	if disabled {
		return s
	}
	return ansiFgBlue + s + ansiFgReset
}

func Magenta(s string) string {
	if disabled {
		return s
	}
	return ansiFgMagenta + s + ansiFgReset
}

func Cyan(s string) string {
	if disabled {
		return s
	}
	return ansiFgCyan + s + ansiFgReset
}

func Underline(s string) string {
	if disabled {
		return s
	}
	return ansiUnderline + s + ansiUnderlineOff
}
