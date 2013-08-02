// Package scrypto implements simple to use cryptographic signatures.
package scrypto

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

func GenerateKey() *Key {
	pk, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if e != nil {
		panic(e)
	}
	return &Key{*pk}
}

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

func (key *Key) Sign(data []byte) (string, error) {
	h := sha256.New()
	hash := h.Sum(data)

	s := signature{}
	var e error
	s.r, s.s, e = ecdsa.Sign(rand.Reader, &key.PrivateKey, hash)
	if e != nil {
		return "", e
	}

	return s.String(), nil
}

func (key *Key) Verify(data []byte, sig string) (bool, error) {
	h := sha256.New()
	hash := h.Sum(data)
	sign, e := signatureFromString(sig)
	if e != nil {
		return false, e
	}

	success := ecdsa.Verify(&key.PublicKey, hash, sign.r, sign.s)
	return success, nil
}

func (k *Key) String() string {
	return slicesToString([][]byte{k.X.Bytes(), k.Y.Bytes(), k.D.Bytes()})
}

func (s *signature) String() string {
	return slicesToString([][]byte{s.r.Bytes(), s.s.Bytes()})
}

func signatureFromString(s string) (*signature, error) {
	sig := signature{}
	ss, e := stringToSlices(s)
	if e != nil {
		return nil, e
	}

	sig.r = big.NewInt(0).SetBytes(ss[0])
	sig.s = big.NewInt(0).SetBytes(ss[1])

	return &sig, nil
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
	ss := make([][]byte, 0)

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

