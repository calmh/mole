package main

import (
	"log"
	"syscall"
)

func init() {
	var rLimit syscall.Rlimit
	rLimit.Max = 4096
	rLimit.Cur = 4096

	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Println("Warning: Setrlimit:", err)
	}
}
