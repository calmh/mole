package main

import (
	"net/http"
)

func init() {
	handlers["/ping"] = handler{ping, true}
}

func ping(rw http.ResponseWriter, req *http.Request) {
}
