package main

import (
	"os"
	"sync"
)

var (
	exitFns  []func()
	exitOnce sync.Once
)

func atExit(fn func()) {
	exitFns = append(exitFns, fn)
}

func exit(code int) {
	exitOnce.Do(func() {
		// Execute exit funcs in reverse order.
		l := len(exitFns) - 1
		for i := range exitFns {
			j := l - i
			debugln("exit handler", j)
			exitFns[j]()
		}
		os.Exit(code)
	})
}
