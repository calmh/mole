package main

import (
	"fmt"
	"github.com/calmh/mole/ticket"
	"github.com/mavricknz/ldap"
	"log"
	"net"
	"net/http"
	"time"
)

func backendAuthenticate(user, password string) bool {
	c := ldap.NewLDAPConnection(ldapServer, uint16(ldapPort))
	err := c.Connect()
	if err != nil {
		log.Println("ldap:", err)
		return false
	}

	err = c.Bind(fmt.Sprintf(bindTemplate, user), password)
	if err != nil {
		log.Printf("ldap: %q: %s", user, err)
		return false
	}

	return true
}

func authenticate(rw http.ResponseWriter, req *http.Request) bool {
	tic := req.Header.Get("X-Mole-Ticket")
	if tic == "" {
		return false
	}

	addr, err := net.ResolveTCPAddr("tcp", req.RemoteAddr)
	if err != nil {
		// Resolving the remote address should never fail
		panic(err)
	}
	ip := addr.IP.String()
	if ip == "" {
		// The remote address should never be empty
		panic("bug: empty remote address")
	}

	user, err := ticket.Verify(tic, ip, time.Now().Unix())
	if err != nil {
		return false
	}

	rw.Header().Set("X-Mole-Authenticated", user)
	req.Header.Set("X-Mole-Authenticated", user)
	return true
}
