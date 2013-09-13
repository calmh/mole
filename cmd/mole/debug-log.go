package main

import (
	"fmt"
	"log"
	"os"
)

var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

func debug(vals ...interface{}) {
	if globalOpts.Debug {
		s := fmt.Sprintln(vals...)
		logger.Output(2, s)
	}
}

func debugf(format string, vals ...interface{}) {
	if globalOpts.Debug {
		s := fmt.Sprintf(format, vals...)
		logger.Output(2, s)
	}
}
