// Package upgrade performs automatic upgrade of the running binary.
package upgrade

import (
	"bitbucket.org/kardianos/osext"
	"compress/gzip"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// Build is a versioned build of a binary.
type Build struct {
	URL        string
	Hash       string
	BuildStamp int
	Version    string
}

// The binary's SHA1 hash was incorrect after downloading.
var ErrHashMismatch = errors.New("hash mismatch")

// Newest checks for the newest build available for a given binary at a given
// base URL and returns a Build on an error.
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

// UpgradeTo upgrades the currently exeuting binary to the specified Build and
// returns an error or nil. The file is downloaded to <destination>.part and
// atomically renamed to the destination after the hash check. The destination
// is taken as the path of the currently executing binary, while following any
// symbolic links to it's destination.
func UpgradeTo(build Build) error {
	path, err := osext.Executable()
	if err != nil {
		return err
	}

	path, err = filepath.EvalSymlinks(path)
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

	ftime := time.Unix(int64(build.BuildStamp), 0)
	_ = os.Chtimes(tmp, ftime, ftime)

	return os.Rename(tmp, path)
}

func sha1file(fname string) (hash string, err error) {
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	defer f.Close()

	h := sha1.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return
	}
	hb := h.Sum(nil)
	hash = fmt.Sprintf("%x", hb)

	return
}
