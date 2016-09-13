package main

import (
	"net/http"
	"path"
)

func init() {
	addHandler(handler{
		pattern: "/store/",
		method:  "GET",
		fn:      getFile,
		auth:    true,
		ro:      true,
	})
}

func getFile(rw http.ResponseWriter, req *http.Request) {
	iniFile := path.Join(storeDir, "data", req.URL.Path[7:])
	http.ServeFile(rw, req, iniFile)
}
