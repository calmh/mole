package configuration_test

import (
	"testing"

	"github.com/calmh/mole/configuration"
)

func TestLoadFileWithoutError(t *testing.T) {
	_, err := configuration.LoadFile("test.ini")
	if err != nil {
		t.Error(err)
	}
}

func TestGeneralSection(t *testing.T) {
	cfg, _ := configuration.LoadFile("test.ini")

	if cfg.General.Description != "Operator (One)" {
		t.Errorf("Incorrect Description %q", cfg.General.Description)
	}
	if cfg.General.Author != "Jakob Borg <jakob@nym.se>" {
		t.Errorf("Incorrect Author %q", cfg.General.Author)
	}
	if cfg.General.Version != 400 {
		t.Errorf("Incorrect Version %d", cfg.General.Version)
	}
	if cfg.General.Main != "tac1" {
		t.Errorf("Incorrect Main %q", cfg.General.Main)
	}

	if l := len(cfg.General.Other); l != 1 {
		t.Errorf("Incorrect len(Other) %d", l)
	}
	if s := cfg.General.Other["unrecognized"]; s != "directive" {
		t.Errorf("Incorrect unrecognized %q", s)
	}
}

func TestHosts(t *testing.T) {
	cfg, _ := configuration.LoadFile("test.ini")

	if l := len(cfg.Hosts); l != 2 {
		t.Errorf("Incorrect len(Hosts) %d", l)
	}

	h := cfg.Hosts[cfg.HostsMap["tac1"]]
	if h.Name != "tac1" {
		t.Errorf("Incorrect Name %q", h.Name)
	}
	if h.Addr != "172.16.32.32" {
		t.Errorf("Incorrect Addr %q", h.Addr)
	}
	if h.Port != 22 {
		t.Errorf("Incorrect Port %d", h.Port)
	}
	if h.User != "mole1" {
		t.Errorf("Incorrect User %q", h.User)
	}
	if h.Key != "test\nkey" {
		t.Errorf("Incorrect Key %q", h.Key)
	}
	if h.Pass != "" {
		t.Errorf("Incorrect Pass %q", h.Pass)
	}
	if h.Prompt != `(%|\$|#|>)\s*$` {
		t.Errorf("Incorrect Prompt %q", h.Prompt)
	}
	if h.Via != "tac2" {
		t.Errorf("Incorrect Via %q", h.Via)
	}
	if l := len(h.Other); l != 2 {
		t.Errorf("Incorrect len(Other) %d", l)
	}

	h = cfg.Hosts[cfg.HostsMap["tac2"]]
	if h.Name != "tac2" {
		t.Errorf("Incorrect Name %q", h.Name)
	}
	if h.Addr != "172.16.32.33" {
		t.Errorf("Incorrect Addr %q", h.Addr)
	}
	if h.Port != 2222 {
		t.Errorf("Incorrect Port %d", h.Port)
	}
	if h.User != "mole2" {
		t.Errorf("Incorrect User %q", h.User)
	}
	if h.Key != "" {
		t.Errorf("Incorrect Key %q", h.Key)
	}
	if h.Pass != "testpass" {
		t.Errorf("Incorrect Pass %q", h.Pass)
	}
	if h.Prompt != "~>" {
		t.Errorf("Incorrect Prompt %q", h.Prompt)
	}
	if h.Via != "" {
		t.Errorf("Incorrect Via %q", h.Via)
	}
	if l := len(h.Other); l != 0 {
		t.Errorf("Incorrect len(Other) %d", l)
	}
}

func TestForwards(t *testing.T) {
	cfg, _ := configuration.LoadFile("test.ini")

	if l := len(cfg.Forwards); l != 2 {
		t.Errorf("Incorrect len(Forwards) %d", l)
	}

	f := cfg.Forwards[0]
	if l := len(f.Lines); l != 2 {
		t.Errorf("Incorrect len(Lines) %d", l)
	}

	l1 := f.Lines[0]
	if l1.SrcIP != "127.0.0.1" {
		t.Errorf("Incorrect SrcIP %q", l1.SrcIP)
	}
	if l1.SrcPort != 42000 {
		t.Errorf("Incorrect SrcPort %d", l1.SrcPort)
	}
	if l1.DstIP != "192.168.173.10" {
		t.Errorf("Incorrect DstIP %q", l1.DstIP)
	}
	if l1.DstPort != 42000 {
		t.Errorf("Incorrect DstPort %d", l1.DstPort)
	}
	if l1.Repeat != 2 {
		t.Errorf("Incorrect Repeat %d", l1.Repeat)
	}

	l2 := f.Lines[1]
	if l2.SrcIP != "127.0.0.1" {
		t.Errorf("Incorrect l2 SrcIP %q", l2.SrcIP)
	}
	if l2.SrcPort != 8443 {
		t.Errorf("Incorrect l2 SrcPort %d", l2.SrcPort)
	}
	if l2.DstIP != "192.168.173.10" {
		t.Errorf("Incorrect l2 DstIP %q", l2.DstIP)
	}
	if l2.DstPort != 443 {
		t.Errorf("Incorrect l2 DstPort %d", l2.DstPort)
	}
	if l2.Repeat != 0 {
		t.Errorf("Incorrect l2 Repeat %d", l2.Repeat)
	}
}
