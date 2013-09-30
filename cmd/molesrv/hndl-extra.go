package main

import (
	"net/http"
)

func init() {
	handlers["/extra/"] = handler{extraFile, false}
}

func extraFile(rw http.ResponseWriter, req *http.Request) {
	http.ServeFile(rw, req, storeDir+"/extra/"+req.URL.Path[7:])
}
