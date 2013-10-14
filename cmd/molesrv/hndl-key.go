package main

import (
	"encoding/json"
	"net/http"
)

func init() {
	addHandler(handler{
		pattern: "/key/",
		method:  "GET",
		fn:      getKey,
		auth:    true,
		ro:      true,
	})
}

func getKey(rw http.ResponseWriter, req *http.Request) {
	if key, ok := keys[req.URL.Path[5:]]; ok {
		bs, _ := json.Marshal(struct {
			Key string `json:"key"`
		}{key})
		rw.Write(bs)
	} else {
		rw.WriteHeader(404)
	}
}
