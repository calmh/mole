// Package ansi provides trivial ANSI formatting of strings.
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

// Disable disables ANSI formatting, effectively turning all formatting
// functions into the identity transform.
func Disable() {
	disabled = true
}

// Strlen returns the length of a string, as it will be displayed, in runes.
func Strlen(s string) int {
	cleaned := ansiRe.ReplaceAllString(s, "")
	return utf8.RuneCountInString(cleaned)
}

// Bold returns the string s with bold formatting.
func Bold(s string) string {
	if disabled {
		return s
	}
	return ansiBold + s + ansiBoldOff
}

// Faint returns the string s with faint formatting.
func Faint(s string) string {
	if disabled {
		return s
	}
	return ansiFaint + s + ansiBoldOff
}

// Black returns the string s with black foreground color.
func Black(s string) string {
	if disabled {
		return s
	}
	return ansiFgBlack + s + ansiFgReset
}

// Red returns the string s with red foreground color.
func Red(s string) string {
	if disabled {
		return s
	}
	return ansiFgRed + s + ansiFgReset
}

// Green returns the string s with green foreground color.
func Green(s string) string {
	if disabled {
		return s
	}
	return ansiFgGreen + s + ansiFgReset
}

// Yellow returns the string s with yellow foreground color.
func Yellow(s string) string {
	if disabled {
		return s
	}
	return ansiFgYellow + s + ansiFgReset
}

// Blue returns the string s with blue foreground color.
func Blue(s string) string {
	if disabled {
		return s
	}
	return ansiFgBlue + s + ansiFgReset
}

// Magenta returns the string s with magenta foreground color.
func Magenta(s string) string {
	if disabled {
		return s
	}
	return ansiFgMagenta + s + ansiFgReset
}

// Cyan returns the string s with cyan foreground color.
func Cyan(s string) string {
	if disabled {
		return s
	}
	return ansiFgCyan + s + ansiFgReset
}

// Underline returns the string s with underlined formatting.
func Underline(s string) string {
	if disabled {
		return s
	}
	return ansiUnderline + s + ansiUnderlineOff
}
