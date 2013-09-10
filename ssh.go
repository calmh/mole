package main

import (
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/calmh/mole/configuration"
)

func ssh(h configuration.Host) {
	// An SSH client is represented with a ClientConn. Currently only
	// the "password" authentication method is supported.
	//
	// To authenticate with the remote server you must pass at least one
	// implementation of ClientAuth via the Auth field in ClientConfig.
	config := &ClientConfig{
		User: h.User,
		Auth: []ClientAuth{
			// ClientAuthPassword wraps a ClientPassword implementation
			// in a type that implements ClientAuth.
			ClientAuthPassword(password(h.Pass)),
		},
	}
	client, err := Dial("tcp", fmt.Sprintf("%s:%d", h.Addr, h.Port), config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("/usr/bin/whoami"); err != nil {
		panic("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())
}
