package main

import (
	"encoding/gob"
	"encoding/json"
	"github.com/calmh/mole/conf"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

func init() {
	addHandler(handler{
		pattern: "/store",
		method:  "GET",
		fn:      storeList,
		auth:    true,
		ro:      true,
	})
}

var listCache []listItem
var listCacheLock sync.Mutex

type listItem struct {
	Name        string
	Description string
	Hosts       []string
	Version     float64
	Features    uint32
}

func storeList(rw http.ResponseWriter, req *http.Request) {
	defer listCacheLock.Unlock()
	listCacheLock.Lock()

	if listCache == nil {
		files, err := filepath.Glob(storeDir + "/data/*.ini")
		if err != nil {
			rw.WriteHeader(500)
			rw.Write([]byte(err.Error()))
			return
		}

		for _, file := range files {
			item := listItem{
				Name: path.Base(file[:len(file)-4]),
			}

			f, err := os.Open(file)
			if err != nil {
				log.Printf("Warning: %q: %s", file, err)
				item.Features = conf.FeatureError
				item.Description = "- unreadable -"
				listCache = append(listCache, item)
				continue
			}

			cfg, err := conf.Load(f)
			f.Close()
			if err != nil {
				log.Printf("Warning: %q: %s", file, err)
				item.Features = conf.FeatureError
				item.Description = "- parse error -"
				listCache = append(listCache, item)
				continue
			}

			var hosts []string
			for _, h := range cfg.Hosts {
				hosts = append(hosts, h.Name)
			}

			item.Features = cfg.FeatureFlags()
			item.Description = cfg.General.Description
			item.Hosts = hosts
			item.Version = float64(cfg.General.Version) / 100
			listCache = append(listCache, item)
		}
	}

	if strings.Contains(req.Header.Get("Accept"), "application/gob") {
		rw.Header().Set("Content-Type", "application/gob")
		enc := gob.NewEncoder(rw)
		enc.Encode(listCache)
	} else {
		bs, _ := json.Marshal(listCache)
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(bs)
	}
}
