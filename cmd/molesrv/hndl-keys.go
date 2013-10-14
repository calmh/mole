package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func init() {
	addHandler(handler{
		pattern: "/keys",
		method:  "POST",
		fn:      getKeys,
		auth:    true,
		ro:      true,
	})
}

func getKeys(rw http.ResponseWriter, req *http.Request) {
	var keylist []string
	var keymap = map[string]string{}

	bs, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	err = json.Unmarshal(bs, &keylist)
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	for _, key := range keylist {
		if secret, ok := keys[key]; ok {
			keymap[key] = secret
		} else {
			rw.WriteHeader(404)
			rw.Write([]byte(key))
			return
		}
	}

	bs, _ = json.Marshal(keymap)
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(bs)
}
