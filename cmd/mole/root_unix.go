// +build darwin linux

package main

import (
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
			fatalln("SUDO_UID", err)
		}
	} else {
		realUid = syscall.Getuid()
	}

	e := syscall.Setreuid(-1, realUid)
	fatalErr(e)
}

func requireRoot(reason string) {
	if !hasRoot {
		fatalf(msgErrGainRoot, reason)
	}
}
