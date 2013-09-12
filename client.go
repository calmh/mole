package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
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

var obfuscatedRe = regexp.MustCompile(`\$mole\$[0-9a-f-]{36}`)

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
		log.Fatal(err)
	}

	if resp.StatusCode != 200 {
		data, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		log.Println(resp.Status)
		log.Fatal(string(data))
	}

	debug(resp.Status, resp.Header.Get("Content-type"), resp.ContentLength)

	return resp
}

func (c *Client) List() []ListItem {
	resp := c.request("GET", "/store")
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var items []ListItem
	json.Unmarshal(data, &items)
	sort.Sort(listItems(items))
	return items
}

func (c *Client) Get(tunnel string) string {
	resp := c.request("GET", "/store/"+tunnel+".ini")
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	res := string(data)

	return res
}

func (c *Client) Deobfuscate(tunnel string) string {
	matches := obfuscatedRe.FindAllString(tunnel, -1)
	for _, o := range matches {
		tunnel = strings.Replace(tunnel, o, c.getToken(o[6:]), -1)
	}

	return tunnel
}

func (c *Client) getToken(token string) string {
	resp := c.request("GET", "/key/"+token)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var res map[string]string
	json.Unmarshal(data, &res)
	return fmt.Sprintf("%q", res["key"])
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
