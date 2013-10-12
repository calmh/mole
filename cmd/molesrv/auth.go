package main

import (
	"github.com/calmh/mole/ticket"
	"net"
	"net/http"
	"time"
)

var authBackends = map[string]func(string, string) bool{
	"none": nil, // The nil backend always succeeds
}

func backendAuthenticate(user, password string) bool {
	fn, ok := authBackends[auth]
	if !ok {
		return false
	}
	if fn == nil {
		return true
	}
	return fn(user, password)
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
