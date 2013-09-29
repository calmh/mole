package ticket

import (
	"bytes"
	"math/rand"
	"testing"
	"testing/quick"
)

func TestEncryptDecryptOK(t *testing.T) {
	check := func(blob []byte) bool {
		enc := hashAndEncrypt(blob)
		dec, err := decryptAndHash(enc)
		if err != nil {
			return false
		}
		if bytes.Compare(blob, dec) != 0 {
			return false
		}
		return true
	}

	err := quick.Check(check, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestEncryptDecryptFail(t *testing.T) {
	check := func(blob []byte) bool {
		enc := hashAndEncrypt(blob)
		enc[rand.Intn(len(enc))]++
		dec, err := decryptAndHash(enc)
		if err == nil {
			return false
		}
		if dec != nil {
			return false
		}
		return true
	}

	err := quick.Check(check, nil)
	if err != nil {
		t.Error(err)
	}
}
