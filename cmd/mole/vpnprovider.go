package main

import (
	"github.com/calmh/mole/conf"
)

type VPN interface {
	Stop()
}

type VPNProvider interface {
	Start(*conf.Config) VPN
}

var vpnProviders = make(map[string]VPNProvider)

func startVpn(provider string, cfg *conf.Config) VPN {
	prov, ok := vpnProviders[provider]
	if !ok {
		fatalf(msgErrNoVPN, provider)
	}
	return prov.Start(cfg)
}
