package main

import (
	"encoding/json"
	"github.com/calmh/mole/conf"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

func init() {
	handlers["/store"] = handler{storeList, true}
}

var listCache []byte

type listItem struct {
	Name        string
	Description string
	Vpnc        bool
	OpenConnect bool
	Socks       bool
	Hosts       []string
	LocalOnly   bool
	Version     float64
}

func storeList(rw http.ResponseWriter, req *http.Request) {
	if listCache == nil {
		files, err := filepath.Glob(storeDir + "/data/*.ini")
		if err != nil {
			rw.WriteHeader(500)
			rw.Write([]byte(err.Error()))
			return
		}

		var items []listItem
		for _, file := range files {
			f, err := os.Open(file)
			if err != nil {
				log.Printf("Warning: skipping %q: %s", file, err)
				continue
			}

			cfg, err := conf.Load(f)
			f.Close()
			if err != nil {
				log.Printf("Warning: skipping %q: %s", file, err)
				continue
			}

			var hosts []string
			for _, h := range cfg.Hosts {
				hosts = append(hosts, h.Name)
			}

			item := listItem{
				Name:        path.Base(file[:len(file)-4]),
				Description: cfg.General.Description,
				Vpnc:        cfg.Vpnc != nil,
				OpenConnect: cfg.OpenConnect != nil,
				Socks:       cfg.General.Main != "" && cfg.Hosts[cfg.HostsMap[cfg.General.Main]].SOCKS != "",
				Hosts:       hosts,
				LocalOnly:   len(hosts) == 0,
				Version:     float64(cfg.General.Version) / 100,
			}
			items = append(items, item)
		}

		listCache, err = json.Marshal(items)
		if err != nil {
			rw.WriteHeader(500)
			rw.Write([]byte(err.Error()))
			return
		}
	}

	rw.Write(listCache)
}
