package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

var clientVersion = strings.Replace(buildVersion, "v", "", 1)

type Client struct {
	Ticket string
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

func certFingerprint(conn *tls.Conn) string {
	cert := conn.ConnectionState().PeerCertificates[0].Raw
	sha := sha1.New()
	sha.Write(cert)
	hash := sha.Sum(nil)
	return fmt.Sprintf("%x", hash)
}

func NewClient(host, fingerprint string) *Client {
	if host == "" {
		fatalln(msgNoHost)
	}

	if !strings.HasPrefix(clientVersion, "4.") {
		// Built from go get, so no tag info
		clientVersion = "4.0-unknown-dev"
	}

	transport := &http.Transport{
		Dial: func(n, a string) (net.Conn, error) {
			tlsCfg := &tls.Config{InsecureSkipVerify: true}
			conn, err := tls.Dial(n, host, tlsCfg)
			if err != nil {
				return nil, err
			}

			fp := certFingerprint(conn)
			if fingerprint != "" && fp != fingerprint {
				return nil, fmt.Errorf("server fingerprint mismatch (%s != %s)", fp, serverIni.fingerprint)
			}
			return conn, err
		},
	}
	client := &http.Client{Transport: transport}
	return &Client{host: host, client: client}
}

func (c *Client) request(method, path string, content io.Reader) (*http.Response, error) {
	url := "http://" + c.host + path
	debugln(method, url)

	req, err := http.NewRequest(method, url, content)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Mole-Version", clientVersion)
	req.Header.Add("X-Mole-Ticket", c.Ticket)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s: %s", resp.Status, data)
	}

	debugln(resp.Status, resp.Header.Get("Content-type"), resp.ContentLength)

	return resp, nil
}

func (c *Client) Ping() (string, error) {
	resp, err := c.request("GET", "/ping", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	return string(data), err
}

func (c *Client) List() ([]ListItem, error) {
	resp, err := c.request("GET", "/store", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var items []ListItem
	err = json.Unmarshal(data, &items)
	if err != nil {
		return nil, err
	}

	for i := range items {
		items[i].IntVersion = int(100 * items[i].Version)
	}
	sort.Sort(listItems(items))

	return items, nil
}

func (c *Client) Get(tunnel string) (string, error) {
	resp, err := c.request("GET", "/store/"+tunnel+".ini", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	res := string(data)

	return res, nil
}

func (c *Client) Put(tunnel string, data io.Reader) error {
	resp, err := c.request("PUT", "/store/"+tunnel+".ini", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *Client) Deobfuscate(tunnel string) (string, error) {
	matches := obfuscatedRe.FindAllString(tunnel, -1)
	for _, o := range matches {
		s, err := c.resolveKey(o[6:])
		if err != nil {
			return "", err
		}
		tunnel = strings.Replace(tunnel, o, s, -1)
	}

	return tunnel, nil
}

func (c *Client) GetTicket(username, password string) (string, error) {
	resp, err := c.request("POST", "/ticket/"+username, bytes.NewBufferString(password))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	res := string(data)

	return res, nil
}

func (c *Client) UpgradesURL() (string, error) {
	resp, err := c.request("GET", "/extra/upgrades.json", nil)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var manifest upgradeManifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return "", err
	}

	return manifest.URL, nil
}

func (c *Client) resolveKey(key string) (string, error) {
	resp, err := c.request("GET", "/key/"+key, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var res map[string]string
	err = json.Unmarshal(data, &res)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%q", res["key"]), nil
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
