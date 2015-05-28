// +build darwin linux

package main

import (
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var tunModulesRe = regexp.MustCompile(`\bfoo\.tun\b|\bnet\.tunnelblick\.tun|\bcom\.viscosityvpn\.Viscosity\.tun\b|\bnet\.sf\.tuntaposx\.tun\b`)

func ensureTunModule() bool {
	if runtime.GOOS != "darwin" {
		// This is only relevant on Darwin
		return true
	}

	loaded := tunModuleLoaded()
	debugln("tunModuleLoaded", loaded)
	if loaded {
		return true
	}

	loadTunModule()
	loaded = tunModuleLoaded()
	debugln("tunModuleLoaded", loaded)
	return loaded
}

func tunModuleLoaded() bool {
	cmd := exec.Command("kextstat", "-kl")
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}

	return tunModulesRe.Match(bs)
}

func loadTunModule() {
	requireRoot("kextload")
	becomeRoot()
	defer dropRoot()
	debugln("kextload", "/Library/Extensions/tun.kext")
	cmd := exec.Command("kextload", "/Library/Extensions/tun.kext")
	bs, err := cmd.CombinedOutput()
	outs := strings.TrimSpace(string(bs))
	debugln(outs)
	if err != nil {
		warnln("kextload:", outs)
	}
}
