// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ldap

import (
	//"encoding/hex"
	"fmt"
	"github.com/mavricknz/asn1-ber"
	"testing"
	//"bytes"
)

var modDNs []string = []string{"cn=bob,o=bigcorp", "uid=eworker,ou=people,o=smallcrop"}
var modAttrs []string = []string{"cn"} //,"uid","sn"}
var modValues []string = []string{"aaa", "bbb", "ccc"}

func TestModify(t *testing.T) {
	for _, dn := range modDNs {
		modreq := NewModifyRequest(dn)
		for _, attr := range modAttrs {
			for _, val := range modValues {
				mod := NewMod(ModAdd, attr, []string{val})
				modreq.AddMod(mod)
			}
		}
		fmt.Print(modreq)
	}
}

func TestModifyMods(t *testing.T) {
	for _, dn := range modDNs {
		mods := make([]Mod, 0, 5)
		modreq := NewModifyRequest(dn)
		for _, attr := range modAttrs {
			mod := NewMod(ModAdd, attr, modValues)
			mods = append(mods, *mod)
		}
		modreq.AddMods(mods)
		fmt.Print(modreq)
	}
}

func TestModifyEncodeModifyRequest(t *testing.T) {
	modreq := NewModifyRequest(modDNs[0])
	mod := NewMod(ModAdd, modAttrs[0], modValues)
	modundo := NewMod(ModDelete, modAttrs[0], modValues)
	modreq.AddMod(mod)
	modreq.AddMod(modundo)
	p := encodeModifyRequest(modreq)
	ber.PrintPacket(p)
}
