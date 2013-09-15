package main

import (
	"os/user"

	"nym.se/mole/ansi"
)

func init() {
	ansi.Disable()
}

func requireRoot(reason string) {
}

func getHomeDir() string {
	user, err := user.Current()
	fatalErr(err)
	return user.HomeDir
}
