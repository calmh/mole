//+build !windows

package main

import (
	"golang.org/x/crypto/ssh/terminal"
)

func isTerminal(fd uintptr) bool {
	return terminal.IsTerminal(int(fd))
}
