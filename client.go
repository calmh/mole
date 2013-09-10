package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
)

const ClientVersion = "3.99"

type Client struct {
	host   string
	client *http.Client
}

type ListItem struct {
	Name        string
	Description string
	Vpnc        bool
	OpenConnect bool
	Socks       bool
	Hosts       []string
	LocalOnly   bool
	Version     int
}

func NewClient(host string, cert tls.Certificate) *Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		},
	}
	client := &http.Client{Transport: transport}
	return &Client{host, client}
}

func (c *Client) request(method, path string) *http.Response {
	url := "https://" + c.host + path
	if globalOpts.Debug {
		log.Println(method, url)
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("X-Mole-Version", ClientVersion)

	resp, err := c.client.Do(req)
	if err != nil {
		panic(err)
	}

	if globalOpts.Debug {
		log.Println(resp.Status, resp.Header.Get("Content-type"), resp.ContentLength)
	}

	return resp
}

func (c *Client) List() []ListItem {
	resp := c.request("GET", "/store")
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var items []ListItem
	json.Unmarshal(data, &items)
	sort.Sort(listItems(items))
	return items
}

type listItems []ListItem

func (l listItems) Len() int {
	return len(l)
}

func (l listItems) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

func (l listItems) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
