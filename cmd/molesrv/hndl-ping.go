package main

import (
	"net/http"
)

func init() {
	addHandler(handler{
		pattern: "/ping",
		method:  "GET",
		fn:      ping,
		auth:    true,
		ro:      true,
	})
}

func ping(rw http.ResponseWriter, req *http.Request) {
}
