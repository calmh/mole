package main

import (
	"fmt"
	"mole/ansi"
	"mole/termsize"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var space = regexp.MustCompile(`\s`)

var prefix map[string]string
var loggingEnabled bool

const (
	indent    = 2   // Indent continuation lines by this many spaces
	maxLength = 128 // Infinitely long lines of text are ugly
	minLength = 50  // There are limits to what sort of shenanigans we put up with
)

func enableLogging() {
	fatalErr = logFatalErr
	loggingEnabled = true
}

func lazySetupPrefixes() {
	if prefix == nil {
		if !loggingEnabled {
			panic("attempting setup before logging enabled")
		}
		prefix = map[string]string{
			"debug": ansi.Magenta("debug "),
			// info has no prefix
			"ok":      ansi.Bold(ansi.Green("ok ")),
			"warning": ansi.Bold(ansi.Yellow("warning ")),
			"fatal":   ansi.Bold(ansi.Red("fatal ")),
		}
	}
}

func writeWrapped(s string, p string) {
	w := termsize.Columns()
	if w > maxLength {
		w = maxLength
	} else if w < minLength {
		w = minLength
	}

	debugPrefix := ""
	if debugEnabled {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			ts := time.Now().Format("15:04:05.000")
			file = path.Base(file)
			debugPrefix = fmt.Sprintf("%s %s:%d ", ts, file, line)
		}
	}

	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	for _, l := range lines {
		writeWrappedLine(debugPrefix+p+l, w)
	}
}

func writeWrappedLine(l string, w int) {
	if len(l) < w {
		os.Stdout.WriteString(l)
	} else {
		words := space.Split(l, -1)
		pos := 0
		for _, word := range words {
			l := ansi.Strlen(word)
			if pos+l >= w-1 { // Leave one empty cell on the right, for aesthetic reasons
				os.Stdout.WriteString("\n" + strings.Repeat(" ", indent))
				pos = indent
			} else if pos > 0 {
				os.Stdout.WriteString(" ")
				pos += 1
			}
			os.Stdout.WriteString(word)
			pos += l
		}
	}
	os.Stdout.WriteString("\n")
}

func debugln(vals ...interface{}) {
	lazySetupPrefixes()
	if debugEnabled {
		s := fmt.Sprintln(vals...)
		writeWrapped(s, prefix["debug"])
	}
}

func debugf(format string, vals ...interface{}) {
	lazySetupPrefixes()
	if debugEnabled {
		s := fmt.Sprintf(format, vals...)
		writeWrapped(s, prefix["debug"])
	}
}

func infoln(vals ...interface{}) {
	lazySetupPrefixes()
	s := fmt.Sprintln(vals...)
	writeWrapped(s, prefix["info"])
}

func infof(format string, vals ...interface{}) {
	lazySetupPrefixes()
	s := fmt.Sprintf(format, vals...)
	writeWrapped(s, prefix["info"])
}

func okln(vals ...interface{}) {
	lazySetupPrefixes()
	s := fmt.Sprintln(vals...)
	writeWrapped(s, prefix["ok"])
}

func okf(format string, vals ...interface{}) {
	lazySetupPrefixes()
	s := fmt.Sprintf(format, vals...)
	writeWrapped(s, prefix["ok"])
}

func warnln(vals ...interface{}) {
	lazySetupPrefixes()
	s := fmt.Sprintln(vals...)
	writeWrapped(s, prefix["warning"])
}

func warnf(format string, vals ...interface{}) {
	lazySetupPrefixes()
	s := fmt.Sprintf(format, vals...)
	writeWrapped(s, prefix["warning"])
}

func fatalln(vals ...interface{}) {
	lazySetupPrefixes()
	s := fmt.Sprintln(vals...)
	writeWrapped(s, prefix["fatal"])
	exit(3)
}

func fatalf(format string, vals ...interface{}) {
	lazySetupPrefixes()
	s := fmt.Sprintf(format, vals...)
	writeWrapped(s, prefix["fatal"])
	exit(3)
}

var fatalErr func(error) = panicFatalErr

func logFatalErr(err error) {
	lazySetupPrefixes()
	if err != nil {
		writeWrapped(err.Error(), prefix["fatal"])
		exit(3)
	}
}

func panicFatalErr(err error) {
	if err != nil {
		panic(err)
	}
}
