package main

import (
	"encoding/json"
	"net/http"
)

func init() {
	handlers["/key/"] = handler{getKey, true}
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
