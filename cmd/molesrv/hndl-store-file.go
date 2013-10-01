package main

import (
	"bytes"
	"github.com/calmh/mole/conf"
	"github.com/calmh/mole/ini"
	"io/ioutil"
	"net/http"
	"os"
)

var obfuscateKeys = []string{
	"key",
	"password",
	"IPSec_secret",
	"Xauth_password",
}

func init() {
	handlers["/store/"] = handler{storeFile, true}
}

func storeFile(rw http.ResponseWriter, req *http.Request) {
	iniFile := storeDir + "/data/" + req.URL.Path[7:]
	if req.Method == "GET" {
		http.ServeFile(rw, req, iniFile)
	} else if req.Method == "PUT" {
		// Read pushed data
		data, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			rw.WriteHeader(500)
			rw.Write([]byte(err.Error()))
			return
		}

		// Verify the configuration
		_, err = conf.Load(bytes.NewBuffer(data))
		if err != nil {
			rw.WriteHeader(500)
			rw.Write([]byte(err.Error()))
			return
		}

		// Get the raw INI
		inf := ini.Parse(bytes.NewBuffer(data))

		// Obfuscate
		shouldSaveKeys := false
		for _, section := range inf.Sections() {
			for _, option := range inf.Options(section) {
				for i := range obfuscateKeys {
					if option == obfuscateKeys[i] {
						val := inf.Get(section, option)
						if oval := obfuscate(val); oval != val {
							inf.Set(section, option, oval)
							shouldSaveKeys = true
						}
						break
					}
				}
			}
		}
		if shouldSaveKeys {
			saveKeys()
		}

		// Save
		outf, err := os.Create(iniFile)
		if err != nil {
			rw.WriteHeader(500)
			rw.Write([]byte(err.Error()))
			return
		}
		inf.Write(outf)
		outf.Close()
	} else {
		rw.WriteHeader(405)
	}
}
