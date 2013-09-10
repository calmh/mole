package main

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
)

func bold(s string) string {
	return ansiBold + s + ansiBoldOff
}

func faint(s string) string {
	return ansiFaint + s + ansiBoldOff
}

func black(s string) string {
	return ansiFgBlack + s + ansiFgReset
}

func red(s string) string {
	return ansiFgRed + s + ansiFgReset
}

func green(s string) string {
	return ansiFgGreen + s + ansiFgReset
}

func yellow(s string) string {
	return ansiFgYellow + s + ansiFgReset
}

func blue(s string) string {
	return ansiFgBlue + s + ansiFgReset
}

func magenta(s string) string {
	return ansiFgMagenta + s + ansiFgReset
}

func cyan(s string) string {
	return ansiFgCyan + s + ansiFgReset
}

func underline(s string) string {
	return ansiUnderline + s + ansiUnderlineOff
}
