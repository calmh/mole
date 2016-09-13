package main

import (
	"fmt"
	"mole/conf"
)

type VPN interface {
	Stop()
}

type VPNProvider interface {
	Start(*conf.Config) (VPN, error)
}

var vpnProviders = make(map[string]VPNProvider)

func startVpn(provider string, cfg *conf.Config) (VPN, error) {
	prov, ok := vpnProviders[provider]
	if !ok {
		return nil, fmt.Errorf(msgErrNoVPN, provider)
	}
	return prov.Start(cfg)
}

func supportsVpn(provider string) bool {
	_, ok := vpnProviders[provider]
	return ok
}
