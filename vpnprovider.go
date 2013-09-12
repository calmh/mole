package main

import (
	"github.com/calmh/mole/configuration"
)

type VPN interface {
	Stop()
}

type VPNProvider interface {
	Start(*configuration.Config) VPN
}

var vpnProviders = make(map[string]VPNProvider)
