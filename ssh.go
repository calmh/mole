package main

import (
	"code.google.com/p/go.crypto/ssh"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/calmh/mole/configuration"
	"io"
	"log"
	"net"
)

type password string

func (p password) Password(user string) (string, error) {
	return string(p), nil
}

type challenge string

func (c challenge) Challenge(user, instruction string, questions []string, echos []bool) ([]string, error) {
	answers := make([]string, len(questions))
	for i := range answers {
		answers[i] = string(c)
	}
	return answers, nil
}

type keyring struct {
	keys []*rsa.PrivateKey
}

func (k *keyring) Key(i int) (interface{}, error) {
	if i < 0 || i >= len(k.keys) {
		return nil, nil
	}
	return &k.keys[i].PublicKey, nil
}

func (k *keyring) Sign(i int, rand io.Reader, data []byte) (sig []byte, err error) {
	hashFunc := crypto.SHA1
	h := hashFunc.New()
	h.Write(data)
	digest := h.Sum(nil)
	return rsa.SignPKCS1v15(rand, k.keys[i], hashFunc, digest)
}

func (k *keyring) loadPEM(data []byte) error {
	block, _ := pem.Decode(data)
	if block == nil {
		return errors.New("ssh: no key found")
	}
	r, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	k.keys = append(k.keys, r)
	return nil
}

func sshVia(conn net.Conn, h configuration.Host) *ssh.ClientConn {
	var auths []ssh.ClientAuth

	if h.Pass != "" {
		auths = append(auths, ssh.ClientAuthPassword(password(h.Pass)))
		auths = append(auths, ssh.ClientAuthKeyboardInteractive(challenge(h.Pass)))
	}
	if h.Key != "" {
		k := &keyring{}
		k.loadPEM([]byte(h.Key))
		auths = append(auths, ssh.ClientAuthKeyring(k))
	}

	config := &ssh.ClientConfig{
		User: h.User,
		Auth: auths,
	}

	var client *ssh.ClientConn
	var err error
	if conn != nil {
		log.Printf(" -> %s@%s", h.User, h.Addr)
		client, err = ssh.Client(conn, config)
	} else {
		log.Printf("SSH %s@%s", h.User, h.Addr)
		client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", h.Addr, h.Port), config)
	}
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	return client
}
