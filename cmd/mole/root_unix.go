// +build darwin linux

package main

import (
	"log"
	"strconv"
	"syscall"
)

var hasRoot bool
var realUid int

func init() {
	hasRoot = syscall.Geteuid() == 0

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

	e := syscall.Setreuid(-1, realUid)
	if e != nil {
		log.Fatal(e)
	}
}

func requireRoot(reason string) {
	if !hasRoot {
		log.Fatalf(msgErrGainRoot, reason)
	}
}
