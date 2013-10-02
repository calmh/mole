package main

import (
	"github.com/calmh/mole/ticket"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	validityPeriod = 86400 * 7 // seconds
)

func init() {
	addHandler(handler{
		pattern: "/ticket/",
		method:  "POST",
		fn:      grantTicket,
		auth:    false,
		ro:      true,
	})
}

func grantTicket(rw http.ResponseWriter, req *http.Request) {
	if len(req.URL.Path) < 8 {
		// Should always include at least "/ticket/"
		panic("bug: path too short")
	}

	user := req.URL.Path[8:]
	if user == "" {
		// Empty username is not permitted
		rw.WriteHeader(401)
		return
	}

	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		// Should have existed a body with a password in it
		rw.WriteHeader(500)
		return
	}

	password := string(bs)
	if password == "" {
		// Empty password is not permitted
		rw.WriteHeader(401)
		return
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

	if !backendAuthenticate(user, password) {
		// Authentication failed
		rw.WriteHeader(401)
		return
	}

	validTo := time.Now().Unix() + validityPeriod
	tic := ticket.Grant(user, ip, validTo)
	rw.Write([]byte(tic))
	return
}
