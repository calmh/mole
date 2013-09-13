// +build darwin linux

package main

import (
	"log"
	"strconv"
	"syscall"
)

var realUid int

func init() {
	sudoUid, ok := syscall.Getenv("SUDO_UID")
	if ok {
		var err error
		realUid, err = strconv.Atoi(sudoUid)
		if err != nil {
			log.Fatal("SUDO_UID", err)
		}
	} else {
		realUid = syscall.Getuid()
	}
}

func gainRoot(reason string) {
	e := syscall.Setreuid(-1, 0)
	if e != nil {
		log.Fatalf(msgErrGainRoot, e, reason)
	}
	debug("euid", syscall.Geteuid())
}

func dropRoot() {
	e := syscall.Setreuid(-1, realUid)
	if e != nil {
		log.Fatal(e)
	}
	debug("euid", syscall.Geteuid())
}
