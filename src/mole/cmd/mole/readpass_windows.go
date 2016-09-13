package main

import (
	"bufio"
	"fmt"
	"os"
)

func readpass(prompt string) string {
	warnln(msgPasswordVisible)
	fmt.Printf(prompt)
	bf := bufio.NewReader(os.Stdin)
	line, _, err := bf.ReadLine()
	if err != nil {
		return ""
	}
	return string(line)
}
