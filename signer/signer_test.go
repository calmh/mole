package signer_test

import (
	"github.com/calmh/mole/signer"
	"math/rand"
	"testing"
	"testing/quick"
)

func TestKeyStringMarshalling(t *testing.T) {
	okey := signer.GenerateKey()
	okeyStr := okey.String()
	nkey := signer.KeyFromString(okeyStr)
	nkeyStr := nkey.String()
	if okeyStr != nkeyStr {
		t.Errorf("key mismatch after marshal/unmarshal; %q, %q", okeyStr, nkeyStr)
	}
}

func TestSignAndVerifyOk(t *testing.T) {
	key := signer.GenerateKey()
	test := func(s []byte) bool {
		sig, _ := key.Sign(s)
		ok, _ := key.Verify(s, sig)
		return ok
	}
	err := quick.Check(test, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestSignAndVerifyFail(t *testing.T) {
	key := signer.GenerateKey()
	test := func(s []byte) bool {
		if len(s) == 0 {
			return true
		}

		sig, _ := key.Sign(s)
		var idx int
		if len(s) > 1 {
			idx = rand.Intn(len(s) - 1)
		}
		s[idx]++
		ok, _ := key.Verify(s, sig)
		return !ok
	}
	err := quick.Check(test, nil)
	if err != nil {
		t.Error(err)
	}
}
