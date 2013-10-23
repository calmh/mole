package main

import (
	"fmt"
	"net"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.net/proxy"
	"github.com/calmh/mole/conf"
)

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

func sshHost(host string, cfg *conf.Config) (*ssh.ClientConn, error) {
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
