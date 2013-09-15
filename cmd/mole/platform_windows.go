package main

import (
	"os/user"
)

func requireRoot(reason string) {
}

func getHomeDir() string {
	user, err := user.Current()
	fatalErr(err)
	return user.HomeDir
}
