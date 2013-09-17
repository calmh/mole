package main

import (
	"nym.se/mole/conf"
	"os/user"

	"nym.se/mole/ansi"
)

func init() {
	ansi.Disable()
}

func requireRoot(reason string) {}

func getHomeDir() string {
	user, err := user.Current()
	fatalErr(err)
	return user.HomeDir
}

func setupHostsFile(tun string, cfg *conf.Config, qualify bool) {}

func restoreHostsFile(tun string, qualify bool) {}
