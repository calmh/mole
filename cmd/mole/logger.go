package main

import (
	"fmt"
	"log"
	"os"

	"nym.se/mole/ansi"
)

var logger = log.New(os.Stdout, "", 0)
var debugConfig bool

func debugln(vals ...interface{}) {
	if globalOpts.Debug {
		if !debugConfig {
			logger.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
			debugConfig = true
		}
		logger.SetPrefix(ansi.Magenta("debug "))
		s := fmt.Sprintln(vals...)
		_ = logger.Output(2, s)
	}
}

func debugf(format string, vals ...interface{}) {
	if globalOpts.Debug {
		if !debugConfig {
			logger.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
			debugConfig = true
		}
		logger.SetPrefix(ansi.Magenta("debug "))
		s := fmt.Sprintf(format, vals...)
		_ = logger.Output(2, s)
	}
}

func infoln(vals ...interface{}) {
	logger.SetPrefix("")
	s := fmt.Sprintln(vals...)
	_ = logger.Output(2, s)
}

func infof(format string, vals ...interface{}) {
	logger.SetPrefix("")
	s := fmt.Sprintf(format, vals...)
	_ = logger.Output(2, s)
}

func okln(vals ...interface{}) {
	logger.SetPrefix(ansi.Bold(ansi.Green("ok ")))
	s := fmt.Sprintln(vals...)
	_ = logger.Output(2, s)
}

func okf(format string, vals ...interface{}) {
	logger.SetPrefix(ansi.Bold(ansi.Green("ok ")))
	s := fmt.Sprintf(format, vals...)
	_ = logger.Output(2, s)
}

func warnln(vals ...interface{}) {
	logger.SetPrefix(ansi.Bold(ansi.Yellow("warning ")))
	s := fmt.Sprintln(vals...)
	_ = logger.Output(2, s)
}

func warnf(format string, vals ...interface{}) {
	logger.SetPrefix(ansi.Bold(ansi.Yellow("warning ")))
	s := fmt.Sprintf(format, vals...)
	_ = logger.Output(2, s)
}

func fatalln(vals ...interface{}) {
	logger.SetPrefix(ansi.Bold(ansi.Red("fatal ")))
	s := fmt.Sprintln(vals...)
	_ = logger.Output(2, s)
	os.Exit(3)
}

func fatalf(format string, vals ...interface{}) {
	logger.SetPrefix(ansi.Bold(ansi.Red("fatal ")))
	s := fmt.Sprintf(format, vals...)
	_ = logger.Output(2, s)
	os.Exit(3)
}

func fatalErr(err error) {
	if err != nil {
		logger.SetPrefix(ansi.Bold(ansi.Red("fatal ")))
		_ = logger.Output(2, err.Error())
		os.Exit(3)
	}
}
