package main

import (
	"encoding/json"
	"github.com/calmh/mole/ticket"
	"net/http"
)

func init() {
	addHandler(handler{
		pattern: "/ticket/",
		method:  "GET",
		fn:      parseTicket,
		auth:    false,
		ro:      true,
	})
}

func parseTicket(rw http.ResponseWriter, req *http.Request) {
	ticStr := req.Header.Get("X-Mole-Ticket")
	tic, err := ticket.Load(ticStr)
	if err != nil {
		rw.WriteHeader(403)
		rw.Write([]byte(err.Error()))
		return
	}

	// Manually create a map with the interesting fields to avoid inadvertently
	// exposing something sensitive such as the Nonce or fields added to the
	// ticket struct in the future.

	exposedFields := map[string]interface{}{
		"user":     tic.User,
		"ips":      tic.IP,
		"validity": tic.Validity,
	}
	json.NewEncoder(rw).Encode(exposedFields)
	return
}
