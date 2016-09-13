package main

import (
	"mole/ticket"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	validityPeriod = 86400 * 7 // seconds
	maxValidIPs    = 4
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

	tic := getTicket(req)
	tic.User = user
	tic.Validity = validTo
	tic.IP = newIPList(tic.IP, ip, maxValidIPs)

	log.Printf("New ticket %q %v %d", tic.User, tic.IP, tic.Validity)
	rw.Write([]byte(tic.String()))
	return
}

func getTicket(req *http.Request) ticket.Ticket {
	ticStr := req.Header.Get("X-Mole-Ticket")
	if ticStr != "" {
		ticp, err := ticket.Load(ticStr)
		if ticp != nil && err == nil {
			return *ticp
		}
	}
	return ticket.Ticket{}
}

func newIPList(ips []string, ip string, max int) []string {
	var newIPs []string
	for _, eip := range ips {
		if eip != ip {
			newIPs = append(newIPs, eip)
		}
	}
	newIPs = append(newIPs, ip)
	if l := len(newIPs); l > max {
		newIPs = newIPs[l-max:]
	}
	return newIPs
}
