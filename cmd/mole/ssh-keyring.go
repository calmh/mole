package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"

	"golang.org/x/crypto/ssh"
)

type keyring struct {
	keys []*rsa.PrivateKey
}

func (k *keyring) Key(i int) (ssh.PublicKey, error) {
	if i < 0 || i >= len(k.keys) {
		return nil, nil
	}
	return ssh.NewPublicKey(&k.keys[i].PublicKey)
}

func (k *keyring) Sign(i int, rand io.Reader, data []byte) (sig []byte, err error) {
	hashFunc := crypto.SHA1
	h := hashFunc.New()
	_, err = h.Write(data)
	if err != nil {
		return nil, err
	}
	digest := h.Sum(nil)
	return rsa.SignPKCS1v15(rand, k.keys[i], hashFunc, digest)
}

func (k *keyring) loadPEM(data []byte) error {
	block, _ := pem.Decode(data)
	if block == nil {
		return errors.New(msgErrPEMNoKey)
	}
	r, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	k.keys = append(k.keys, r)
	return nil
}
