package signer

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"math/big"
)

type Key struct {
	ecdsa.PrivateKey
}

type signature struct {
	r, s *big.Int
}

// GenerateKey returns a new Key for use in signing/verifying messages.
func GenerateKey() *Key {
	pk, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if e != nil {
		panic(e)
	}
	return &Key{*pk}
}

// KeyFromString returns the Key resulting from parsing the string pks, or nil
// if it could not be parsed as a key.
func KeyFromString(pks string) *Key {
	var pk ecdsa.PrivateKey
	pk.Curve = elliptic.P256()

	ss, e := stringToSlices(pks)
	if e != nil {
		return nil
	}

	pk.X = big.NewInt(0).SetBytes(ss[0])
	pk.Y = big.NewInt(0).SetBytes(ss[1])
	pk.D = big.NewInt(0).SetBytes(ss[2])

	return &Key{pk}
}

// Sign returns the signature of the data or an error.
func (key *Key) Sign(data []byte) (string, error) {
	h := sha256.New()
	_, e := h.Write(data)
	if e != nil {
		return "", e
	}
	hash := h.Sum(nil)

	s := signature{}
	s.r, s.s, e = ecdsa.Sign(rand.Reader, &key.PrivateKey, hash)
	if e != nil {
		return "", e
	}

	return signatureToString(s), nil
}

// String is an ASCII serialization of the key.
func (k *Key) String() string {
	return slicesToString([][]byte{k.X.Bytes(), k.Y.Bytes(), k.D.Bytes()})
}

// Verify verifies the signature for the data and returns true/false or an error.
func (key *Key) Verify(data []byte, sig string) (bool, error) {
	h := sha256.New()
	_, e := h.Write(data)
	if e != nil {
		return false, e
	}
	hash := h.Sum(nil)

	sign, e := stringToSignature(sig)
	if e != nil {
		return false, e
	}

	success := ecdsa.Verify(&key.PublicKey, hash, sign.r, sign.s)
	return success, nil
}

func signatureToString(s signature) string {
	return slicesToString([][]byte{s.r.Bytes(), s.s.Bytes()})
}

func stringToSignature(s string) (signature, error) {
	sig := signature{}
	ss, e := stringToSlices(s)
	if e != nil {
		return sig, e
	}

	sig.r = big.NewInt(0).SetBytes(ss[0])
	sig.s = big.NewInt(0).SetBytes(ss[1])

	return sig, nil
}

func slicesToString(ss [][]byte) string {
	var b bytes.Buffer
	for _, s := range ss {
		binary.Write(&b, binary.BigEndian, uint16(len(s)))
		b.Write(s)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func stringToSlices(s string) ([][]byte, error) {
	bs, e := base64.StdEncoding.DecodeString(s)
	if e != nil {
		return nil, e
	}

	b := bytes.NewBuffer(bs)
	var ss [][]byte

	for b.Len() > 0 {
		var l uint16
		binary.Read(b, binary.BigEndian, &l)
		s := make([]byte, l)
		_, e := b.Read(s)
		if e != nil {
			return nil, e
		}
		ss = append(ss, s)
	}

	return ss, nil
}
