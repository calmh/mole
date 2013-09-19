package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
)

const (
	ServerCertificateName = "server"
)

var clientVersion = strings.Replace(buildVersion, "v", "", 1)

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
	Version     float64
	IntVersion  int
}

type upgradeManifest struct {
	URL string
}

var obfuscatedRe = regexp.MustCompile(`\$mole\$[0-9a-f-]{36}`)

func caCert() *x509.Certificate {
	file, err := os.Open(path.Join(globalOpts.Home, "ca-cert.pem"))
	if err != nil {
		return nil
	}
	pemdata, err := ioutil.ReadAll(file)
	if err != nil {
		return nil
	}

	block, _ := pem.Decode(pemdata)
	if block.Type != "CERTIFICATE" {
		return nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil
	}

	return cert
}

func NewClient(host string, cert tls.Certificate) *Client {
	if !strings.HasPrefix(clientVersion, "4.") {
		// Built from go get, so no tag info
		clientVersion = "4.0-unknown-dev"
	}

	crt := caCert()
	if crt != nil {
		pool := x509.NewCertPool()
		pool.AddCert(crt)
		transport := &http.Transport{
			Dial: func(n, a string) (net.Conn, error) {
				return net.Dial(n, host)
			},
			TLSClientConfig: &tls.Config{
				ServerName:   ServerCertificateName,
				RootCAs:      pool,
				Certificates: []tls.Certificate{cert},
			},
		}
		client := &http.Client{Transport: transport}
		return &Client{ServerCertificateName, client}
	} else {
		debugln(msgWarnNoCert)
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				Certificates:       []tls.Certificate{cert},
			},
		}
		client := &http.Client{Transport: transport}
		return &Client{host, client}
	}
}

func (c *Client) request(method, path string, content io.Reader) *http.Response {
	url := "https://" + c.host + path
	debugln(method, url)

	req, err := http.NewRequest(method, url, content)
	fatalErr(err)
	req.Header.Add("X-Mole-Version", clientVersion)

	resp, err := c.client.Do(req)
	fatalErr(err)

	if resp.StatusCode != 200 {
		data, _ := ioutil.ReadAll(resp.Body)
		warnln(resp.Status)
		fatalln(string(data))
	}

	debugln(resp.Status, resp.Header.Get("Content-type"), resp.ContentLength)

	return resp
}

func (c *Client) List() []ListItem {
	resp := c.request("GET", "/store", nil)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	fatalErr(err)
	var items []ListItem
	err = json.Unmarshal(data, &items)
	fatalErr(err)
	for i := range items {
		items[i].IntVersion = int(100 * items[i].Version)
	}
	sort.Sort(listItems(items))
	return items
}

func (c *Client) Get(tunnel string) string {
	resp := c.request("GET", "/store/"+tunnel+".ini", nil)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	fatalErr(err)
	res := string(data)

	return res
}

func (c *Client) Put(tunnel string, data io.Reader) {
	resp := c.request("PUT", "/store/"+tunnel+".ini", data)
	defer resp.Body.Close()
}

func (c *Client) Deobfuscate(tunnel string) string {
	matches := obfuscatedRe.FindAllString(tunnel, -1)
	for _, o := range matches {
		tunnel = strings.Replace(tunnel, o, c.resolveKey(o[6:]), -1)
	}

	return tunnel
}

func (c *Client) UpgradesURL() string {
	url := "https://" + c.host + "/extra/upgrades.json"
	debugln("GET", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	req.Header.Add("X-Mole-Version", clientVersion)

	resp, err := c.client.Do(req)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var manifest upgradeManifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return ""
	}
	return manifest.URL
}

func (c *Client) resolveKey(key string) string {
	resp := c.request("GET", "/key/"+key, nil)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	fatalErr(err)

	var res map[string]string
	err = json.Unmarshal(data, &res)
	fatalErr(err)
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
