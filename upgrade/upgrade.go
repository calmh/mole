package upgrade

import (
	"compress/gzip"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"bitbucket.org/kardianos/osext"
)

type Build struct {
	URL        string
	Hash       string
	BuildStamp int
	Version    string
}

var ErrHashMismatch = errors.New("hash mismatch")

func Newest(binary string, url string) (b Build, err error) {
	archBin := binary + "-" + runtime.GOOS + "-" + runtime.GOARCH
	binUrl := url + "/" + archBin
	resp, err := http.Get(binUrl + ".json")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	b.URL = binUrl
	err = json.Unmarshal(data, &b)
	return
}

func UpgradeTo(build Build) error {
	path, err := osext.Executable()
	if err != nil {
		return err
	}

	tmp := path + ".part"
	gzUrl := build.URL + ".gz"

	resp, err := http.Get(gzUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	gunzip, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}

	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer os.Remove(tmp)

	err = os.Chmod(tmp, 0755)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, gunzip)
	if err != nil {
		return err
	}

	err = out.Close()
	if err != nil {
		return err
	}

	hash, err := sha1file(tmp)
	if err != nil {
		return err
	}

	if hash != build.Hash {
		return ErrHashMismatch
	}

	return os.Rename(tmp, path)
}

func sha1file(fname string) (hash string, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	defer f.Close()

	h := sha1.New()
	io.Copy(h, f)
	hb := h.Sum(nil)
	hash = fmt.Sprintf("%x", hb)

	return
}
