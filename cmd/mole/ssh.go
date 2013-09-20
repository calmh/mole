package main

import (
	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.net/proxy"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/calmh/mole/conf"
	"io"
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

func (k *keyring) Key(i int) (ssh.PublicKey, error) {
	if i < 0 || i >= len(k.keys) {
		return nil, nil
	}
	return ssh.NewRSAPublicKey(&k.keys[i].PublicKey), nil
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

func sshOnConn(conn net.Conn, h conf.Host) (*ssh.ClientConn, error) {
	var auths []ssh.ClientAuth

	if h.Pass != "" {
		auths = append(auths, ssh.ClientAuthPassword(password(h.Pass)))
		auths = append(auths, ssh.ClientAuthKeyboardInteractive(challenge(h.Pass)))
	}

	if h.Key != "" {
		k := &keyring{}
		err := k.loadPEM([]byte(h.Key))
		if err != nil {
			return nil, err
		}
		auths = append(auths, ssh.ClientAuthKeyring(k))
	}

	config := &ssh.ClientConfig{
		User: h.User,
		Auth: auths,
	}

	debugln("handshake & authenticate")
	client, err := ssh.Client(conn, config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func sshHost(host string, cfg *conf.Config) (Dialer, error) {
	h := cfg.Hosts[cfg.HostsMap[host]]
	var conn net.Conn
	var err error
	if h.Via != "" {
		debugln("via", h.Via)
		dialer, err := sshHost(h.Via, cfg)
		if err != nil {
			return nil, err
		}
		dst := fmt.Sprintf("%s:%d", h.Addr, h.Port)
		debugln("dial", dst)
		conn, err = dialer.Dial("tcp", dst)
		if err != nil {
			return nil, err
		}
	} else {
		var dialer Dialer = proxy.Direct
		if h.SOCKS != "" {
			debugln("socks via", h.SOCKS)
			dialer, err = proxy.SOCKS5("tcp", h.SOCKS, nil, proxy.Direct)
			if err != nil {
				return nil, err
			}
		}
		dst := fmt.Sprintf("%s:%d", h.Addr, h.Port)
		debugln("dial", dst)
		conn, err = dialer.Dial("tcp", dst)
	}
	if err != nil {
		return nil, err
	}
	return sshOnConn(conn, h)
}
