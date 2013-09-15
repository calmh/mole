package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"nym.se/mole/ini"
)

var errParams = errors.New("incorrect command line parameters")

var (
	buildVersion string
	buildStamp   string
	buildDate    time.Time
	buildUser    string
	homeDir      string
	serverAddr   string
)

var globalOpts struct {
	Debug bool   `short:"d" long:"debug" description:"Show debug output"`
	Home  string `short:"h" long:"home" description:"Mole home directory" default:"~/.mole" value-name:"DIR"`
	Remap bool   `long:"remap-lo" description:"Use port remapping for extended lo addresses (required/default on Windows)"`
}

var globalParser = flags.NewParser(&globalOpts, flags.Default)

func main() {
	epoch, e := strconv.ParseInt(buildStamp, 10, 64)
	if e == nil {
		buildDate = time.Unix(epoch, 0)
	}

	// TIME LIMITED BETA
	// 14 days self destruct
	if !buildDate.IsZero() && time.Since(buildDate) > 14*24*time.Hour {
		fatalln("This is an expired beta version.\nPlease grab a new build from http://ps-build1.vbg.se.prnw.net/job/mole")
	}
	// TIME LIMITED BETA
	// 14 days self destruct

	if runtime.GOOS == "windows" {
		globalOpts.Remap = true
	}

	globalParser.ApplicationName = "mole"
	if _, e := globalParser.Parse(); e != nil {
		os.Exit(1)
	}

	printTotalStats()
	okln("done")
}

func formatBytes(n uint64) string {
	if n < 1024 {
		return fmt.Sprintf("%d ", n)
	}

	prefixes := []string{" k", " M", " G", " T"}
	divisor := 1024.0
	for i := range prefixes {
		rem := float64(n) / divisor
		if rem < 1024.0 || i == len(prefixes)-1 {
			return fmt.Sprintf("%.02f%s", rem, prefixes[i])
		}
		divisor *= 1024
	}
	return ""
}

func setup() {
	if globalOpts.Debug {
		printVersion()
	}

	userHome := os.Getenv("HOME")
	debugln("userHome", userHome)
	if userHome == "" {
		fatalln(msgErrNoHome)
	}

	if strings.HasPrefix(globalOpts.Home, "~/") {
		homeDir = strings.Replace(globalOpts.Home, "~", userHome, 1)
	}
	debugln("homeDir", homeDir)

	configFile := path.Join(homeDir, "mole.ini")
	f, e := os.Open(configFile)
	fatalErr(e)

	config := ini.Parse(f)
	serverAddr = config.Sections["server"]["host"] + ":" + config.Sections["server"]["port"]
}

func printVersion() {
	infof("mole (%s-%s)", runtime.GOOS, runtime.GOARCH)
	if buildVersion != "" {
		infof("  %s", buildVersion)
	}
	if !buildDate.IsZero() {
		infof("  %v by %s", buildDate, buildUser)
	}
}

func certificate() tls.Certificate {
	cert, err := tls.LoadX509KeyPair(path.Join(homeDir, "mole.crt"), path.Join(homeDir, "mole.key"))
	fatalErr(err)
	return cert
}
