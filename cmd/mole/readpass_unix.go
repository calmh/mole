// +build linux darwin

package main

import (
	"golang.org/x/crypto/ssh/terminal"
	"fmt"
	"os"
)

func readpass(prompt string) string {
	fmt.Printf(prompt)
	bs, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return ""
	}
	fmt.Println()
	return string(bs)
}
