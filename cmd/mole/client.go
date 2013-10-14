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
	"time"
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
	Hosts       []string
	Version     float64
	Features    uint32
	IntVersion  int
}

type upgradeManifest struct {
	URL string
}

var obfuscatedRe = regexp.MustCompile(`\$mole\$[0-9a-zA-Z+/-]+`)

func certFingerprint(conn *tls.Conn) []byte {
	cert := conn.ConnectionState().PeerCertificates[0].Raw
	sha := sha1.New()
	_, _ = sha.Write(cert)
	return sha.Sum(nil)
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

			fp := hexBytes(certFingerprint(conn))
			if fingerprint != "" && fp != fingerprint {
				return nil, fmt.Errorf("server fingerprint mismatch (%s != %s)", fp, fingerprint)
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

	req.Header.Set("User-Agent", "mole/"+clientVersion)
	req.Header.Set("X-Mole-Version", clientVersion)
	req.Header.Set("X-Mole-Ticket", c.Ticket)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 530 {
		defer resp.Body.Close()
		return nil, fmt.Errorf(msg530)
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		data, _ := ioutil.ReadAll(resp.Body)
		if len(data) > 0 {
			return nil, fmt.Errorf("%s: %s", resp.Status, data)
		} else {
			return nil, fmt.Errorf(resp.Status)
		}
	}

	debugln(resp.Status, resp.Header.Get("Content-type"), resp.ContentLength)

	if ch := resp.Header.Get("X-Mole-Canonical-Hostname"); ch != "" && ch != moleIni.Get("server", "host") {
		moleIni.Set("server", "host", ch)
		saveMoleIni()
		okf(msgUpdatedHost, ch)
	}

	return resp, nil
}

func (c *Client) Ping() (string, error) {
	t0 := time.Now()

	resp, err := c.request("GET", "/ping", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	debugf("ping %.01f ms", time.Since(t0).Seconds()*1000)
	return string(data), err
}

func (c *Client) ServerVersion() string {
	url := "http://" + c.host + "/ping"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("User-Agent", "mole/"+clientVersion)
	req.Header.Set("X-Mole-Version", clientVersion)
	req.Header.Set("X-Mole-Ticket", c.Ticket)

	resp, err := c.client.Do(req)
	if err != nil {
		return ""
	}
	resp.Body.Close()

	return resp.Header.Get("X-Mole-Version")
}

func (c *Client) List() ([]ListItem, error) {
	t0 := time.Now()

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

	debugf("list %.01f ms", time.Since(t0).Seconds()*1000)
	return items, nil
}

func (c *Client) Get(tunnel string) (string, error) {
	t0 := time.Now()

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

	debugf("get %.01f ms", time.Since(t0).Seconds()*1000)
	return res, nil
}

func (c *Client) Put(tunnel string, data io.Reader) error {
	t0 := time.Now()

	resp, err := c.request("PUT", "/store/"+tunnel+".ini", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	debugf("put %.01f ms", time.Since(t0).Seconds()*1000)
	return nil
}

func (c *Client) Delete(tunnel string) error {
	t0 := time.Now()

	resp, err := c.request("DELETE", "/store/"+tunnel+".ini", nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	debugf("delete %.01f ms", time.Since(t0).Seconds()*1000)
	return nil
}

func (c *Client) Deobfuscate(tunnel string) (string, error) {
	t0 := time.Now()

	var err error
	var keylist []string
	var keymap map[string]string

	matches := obfuscatedRe.FindAllString(tunnel, -1)
	for _, o := range matches {
		keylist = append(keylist, o[6:])
	}

	bs, _ := json.Marshal(keylist)
	resp, err := c.request("POST", "/keys", bytes.NewBuffer(bs))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bs, err = ioutil.ReadAll(resp.Body)
	fatalErr(err)

	err = json.Unmarshal(bs, &keymap)
	fatalErr(err)

	for k, v := range keymap {
		tunnel = strings.Replace(tunnel, "$mole$"+k, fmt.Sprintf("%q", v), -1)
	}

	debugf("deobfuscate %.01f ms", time.Since(t0).Seconds()*1000)
	return tunnel, nil
}

func (c *Client) GetTicket(username, password string) (string, error) {
	t0 := time.Now()

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

	debugf("getticket %.01f ms", time.Since(t0).Seconds()*1000)
	return res, nil
}

func (c *Client) UpgradesURL() (string, error) {
	t0 := time.Now()

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

	debugf("upgradeurl %.01f ms", time.Since(t0).Seconds()*1000)
	return manifest.URL, nil
}

type Package struct {
	Package     string
	Description string
}

func (c *Client) Packages() (map[string][]Package, error) {
	t0 := time.Now()

	resp, err := c.request("GET", "/extra/packages.json", nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var m map[string][]Package
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	debugf("packages %.01f ms", time.Since(t0).Seconds()*1000)
	return m, nil
}

func (c *Client) Package(file string) (io.ReadCloser, error) {
	resp, err := c.request("GET", "/extra/"+file, nil)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
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
