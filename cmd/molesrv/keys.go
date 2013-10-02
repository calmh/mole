package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path"
	"strings"
)

var keys map[string]string

func loadKeys() error {
	file := path.Join(storeDir, "data", "keys.json")

	fd, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fd.Close()

	d := json.NewDecoder(fd)
	return d.Decode(&keys)
}

func saveKeys() error {
	file := path.Join(storeDir, "data", "keys.json")

	bs, err := json.Marshal(keys)
	if err != nil {
		return err
	}
	var ibs bytes.Buffer
	json.Indent(&ibs, bs, "", "  ")

	fd, err := os.Create(file + ".part")
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = ibs.WriteTo(fd)
	if err != nil {
		return err
	}
	err = fd.Sync()
	if err != nil {
		return err
	}

	return os.Rename(file+".part", file)
}

func randomKey() string {
	keybs := make([]byte, 15)
	_, err := rand.Read(keybs)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(keybs)
}

func obfuscate(val string) string {
	if strings.HasPrefix(val, "$mole$") {
		return val
	} else {
		key := randomKey()
		if _, ok := keys[key]; ok {
			panic("randomly generated key already exists")
		}

		keys[key] = val

		return "$mole$" + key
	}
}
