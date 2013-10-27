//+build !windows

package main

import (
	"code.google.com/p/go.crypto/ssh/terminal"
)

func isTerminal(fd uintptr) bool {
	return terminal.IsTerminal(int(fd))
}
