// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ldap

import (
	"fmt"
	"testing"
)

/*
	type ControlVlvRequest struct {
		Criticality        bool
		BeforeCount        int32
		AfterCount         int32
		ByOffset           []int32
		GreaterThanOrEqual string
		ContextID          []byte
	}
*/
func TestVlvRequest(t *testing.T) {
	fmt.Println("TestVlvRequest starting...")
	vlv := new(ControlVlvRequest)
	vlv.BeforeCount = 0
	vlv.AfterCount = 10

	offset := new(VlvOffSet)
	offset.Offset = 1
	offset.ContentCount = 10

	vlv.ByOffset = offset

	VlvDebug = true
	p, err := vlv.Encode()
	if err != nil {
		t.Error(err)
	}
	// ber.PrintPacket(p)
	// ber.PrintBytes(p)
	if p != nil {
		fmt.Printf("")
	}
	fmt.Println("TestVlvRequest finsished.")
}
