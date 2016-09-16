package main

import (
	"fmt"
	"net"

	"mole/conf"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"
)

func sshOnConn(conn net.Conn, h conf.Host) (*ssh.Client, error) {
	var auths []ssh.AuthMethod

	if h.Pass != "" {
		auths = append(auths, ssh.Password(h.Pass))
		auths = append(auths, ssh.KeyboardInteractive(kbdInteractive(h.Pass)))
	}

	if h.Key != "" {
		k := &keyring{}
		err := k.loadPEM([]byte(h.Key))
		if err != nil {
			return nil, err
		}
		for _, k := range k.keys {
			s, _ := ssh.NewSignerFromKey(k)
			auths = append(auths, ssh.PublicKeys(s))
		}
	}

	config := &ssh.ClientConfig{
		User: h.User,
		Auth: auths,
	}

	debugln("handshake & authenticate")
	cc, nc, reqs, err := ssh.NewClientConn(conn, conn.RemoteAddr().String(), config)
	if err != nil {
		return nil, err
	}
	client := ssh.NewClient(cc, nc, reqs)
	return client, nil
}

func sshHost(host string, cfg *conf.Config) (*ssh.Client, error) {
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

func kbdInteractive(secret string) ssh.KeyboardInteractiveChallenge {
	return func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
		if len(questions) == 0 {
			return nil, nil
		}
		if len(questions) != 1 {
			return nil, fmt.Errorf("too complex questionnaire: %#v", questions)
		}
		return []string{secret}, nil
	}
}
