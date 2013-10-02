package main

import (
	"net/http"
)

func init() {
	addHandler(handler{
		pattern: "/extra/",
		method:  "GET",
		fn:      extraFile,
		auth:    false,
		ro:      true,
	})
}

func extraFile(rw http.ResponseWriter, req *http.Request) {
	http.ServeFile(rw, req, storeDir+"/extra/"+req.URL.Path[7:])
}
