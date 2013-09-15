// +build darwin linux

package ansi

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

func Bold(s string) string {
	return ansiBold + s + ansiBoldOff
}

func Faint(s string) string {
	return ansiFaint + s + ansiBoldOff
}

func Black(s string) string {
	return ansiFgBlack + s + ansiFgReset
}

func Red(s string) string {
	return ansiFgRed + s + ansiFgReset
}

func Green(s string) string {
	return ansiFgGreen + s + ansiFgReset
}

func Yellow(s string) string {
	return ansiFgYellow + s + ansiFgReset
}

func Blue(s string) string {
	return ansiFgBlue + s + ansiFgReset
}

func Magenta(s string) string {
	return ansiFgMagenta + s + ansiFgReset
}

func Cyan(s string) string {
	return ansiFgCyan + s + ansiFgReset
}

func Underline(s string) string {
	return ansiUnderline + s + ansiUnderlineOff
}
