package conf_test

import (
	"github.com/calmh/mole/conf"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func loadFile(f string) (*conf.Config, error) {
	fd, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return conf.Load(fd)
}

type TestCase struct {
	FileGlob   string
	ErrorMatch string
}

var validationCases = []TestCase{
	{"valid-*.ini", ""}, // All valid files are valid
	{"inv-*.ini", "."},  // All invalid files are invalid; more specific checks below
	{"inv-nover.ini", `missing required field "version"`},
	{"inv-nodescr.ini", `missing required field "description"`},
	{"inv-noauthor.ini", `missing required field "author"`},
	{"inv-unknown.ini", `unrecognized field "unrecognized"`},
	{"inv-nohosts-nofwds.ini", `either "hosts" or "forwards"`},
	{"inv-nosuchmain.ini", `nonexistent host "foo"`},
	{"inv-nosuchvia.ini", `nonexistent host "tac2"`},
	{"inv-unknhostattrs.ini", `unrecognized field "foo"`},
	{"inv-fwdcomment.ini", `forward comments`},
	{"inv-duplfwd.ini", `duplicate forward source "127.0.0.1:8443"`},
	{"inv-lowport.ini", `privileged source port 443`},
	{"inv-nohostaddr.ini", `required field "addr"`},
	{"inv-nohostuser.ini", `required field "user"`},
	{"inv-nohostpasskey.ini", `required field "password" or "key"`},
	{"inv-badfwd*.ini", `malformed forward`},
	{"inv-socksvia.ini", `"socks" and "via"`},
}

func TestValidations(t *testing.T) {
	for _, tc := range validationCases {
		fs, err := filepath.Glob("test/" + tc.FileGlob)
		if err != nil {
			panic(err)
		}
		if len(fs) == 0 {
			t.Errorf("test pattern %q matches no files", tc.FileGlob)
		}
		for _, f := range fs {
			cfg, err := loadFile(f)
			if tc.ErrorMatch == "" {
				if cfg == nil {
					t.Errorf("unexpected nil cfg for %s", f)
				}
				if err != nil {
					t.Error(err)
				}
			} else {
				if cfg != nil {
					t.Errorf("unexpected non-nil cfg for %s", f)
				}
				if err == nil {
					t.Errorf("unexpected nil error for %s", f)
				} else if !regexp.MustCompile(tc.ErrorMatch).MatchString(err.Error()) {
					t.Errorf("error %q for %s doesn't match %q", err.Error(), f, tc.ErrorMatch)
				}
			}
		}
	}
}

func TestGeneralSection(t *testing.T) {
	cfg, _ := loadFile("test/valid-general.ini")

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
	cfg, _ := loadFile("test/valid-hosts.ini")

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
	if h.Via != "" {
		t.Errorf("Incorrect Via %q", h.Via)
	}
	if l := len(h.Other); l != 0 {
		t.Errorf("Incorrect len(Other) %d", l)
	}
}

func TestForwards(t *testing.T) {
	cfg, _ := loadFile("test/valid-forwards.ini")

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

	f = cfg.Forwards[1]
	if f.Comments[0] != "yo" {
		t.Errorf("Incorrect f[1] Comment %q", f.Comments[0])
	}
}

func TestSourceAddresses(t *testing.T) {
	cfg, _ := loadFile("test/valid-sourceaddr.ini")

	addrs := cfg.SourceAddresses()
	if l := len(addrs); l != 4 {
		t.Errorf("incorrect len(addrs) %d", l)
	}

	for i, exp := range []string{"127.0.0.12", "127.0.0.13", "127.22.0.16", "127.22.0.17"} {
		if addrs[i] != exp {
			t.Errorf("incorrect addrs[%d] %q != %q", i, addrs[i], exp)
		}
	}
}

func TestVpnc(t *testing.T) {
	cfg, _ := loadFile("test/valid-vpnc.ini")

	if cfg.Vpnc["IPSec_secret"] != "s3cr3t" {
		t.Error("incorrectly parsed vpnc IPSec_secret")
	}
	if cfg.Vpnc["Xauth_password"] != "3v3nm0r3s3cr3t" {
		t.Error("incorrectly parsed vpnc Xauth_password")
	}
}

func TestVpnRoutes(t *testing.T) {
	cfg, _ := loadFile("test/valid-vpnc.ini")

	if l := len(cfg.VpnRoutes); l != 8 {
		t.Errorf("incorrect number of vpn routes %d", l)
	}
	if r := cfg.VpnRoutes[0]; r != "192.168.10.0/24" {
		t.Errorf("incorrect first route %q", r)
	}
}

func TestOpenConnect(t *testing.T) {
	cfg, _ := loadFile("test/valid-openconnect.ini")

	if cfg.OpenConnect["server"] != "foo.example.com" {
		t.Error("incorrectly parsed openconnect server")
	}
}

func TestComments(t *testing.T) {
	cfg, _ := loadFile("test/valid-comments.ini")

	if l := len(cfg.Comments); l != 1 {
		t.Errorf("incorrect file comments len %d", l)
	} else {
		if c := cfg.Comments[0]; c != "Very general comments here" {
			t.Errorf("incorrect general comment %q", c)
		}
	}

	if l := len(cfg.General.Comments); l != 1 {
		t.Errorf("incorrect general comments len %d", l)
	} else {
		if c := cfg.General.Comments[0]; c != "Some general comments" {
			t.Errorf("incorrect general comment %q", c)
		}
	}

	h := cfg.Hosts[0]
	if l := len(h.Comments); l != 2 {
		t.Errorf("incorrect tac1 comments len %d", l)
	} else {
		if c := h.Comments[0]; c != "tac1 comments" {
			t.Errorf("incorrect tac1 comment[0] %q", c)
		}
		if c := h.Comments[1]; c != "further comments" {
			t.Errorf("incorrect tac1 comment[1] %q", c)
		}
	}

	f := cfg.Forwards[0]
	if l := len(f.Comments); l != 2 {
		t.Errorf("incorrect forward[0] comments len %d", l)
	} else {
		if c := f.Comments[0]; c != "This is the residential host" {
			t.Errorf("incorrect forward[0] comment[0] %q", c)
		}
		if c := f.Comments[1]; c != "" {
			t.Errorf("incorrect forward[0] comment[1] %q", c)
		}
	}

	f = cfg.Forwards[1]
	if l := len(f.Comments); l != 2 {
		t.Errorf("incorrect forward[1] comments len %d", l)
	} else {
		// Value from comment= comes before actual comments
		if c := f.Comments[0]; c != "yo" {
			t.Errorf("incorrect forward[1] comment[0] %q", c)
		}
		if c := f.Comments[1]; c != "This is corporate" {
			t.Errorf("incorrect forward[1] comment[1] %q", c)
		}
	}
}
