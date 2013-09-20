package main

import (
	"crypto/tls"
	"fmt"
	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/ini"
	"github.com/calmh/mole/upgrade"
	"github.com/jessevdk/go-flags"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	buildVersion string
	buildStamp   string
	buildDate    time.Time
	buildUser    string
)

var globalOpts struct {
	Home   string `short:"h" long:"home" description:"Mole home directory" default:"~/.mole" value-name:"DIR"`
	Debug  bool   `short:"d" long:"debug" description:"Show debug output"`
	NoAnsi bool   `long:"no-ansi" description:"Disable ANSI formatting sequences"`
	Remap  bool   `long:"remap-lo" description:"Use port remapping for extended lo addresses (required/default on Windows)"`
}

var serverIni struct {
	address     string
	upgrades    bool
	fingerprint string
}

var globalParser = flags.NewParser(&globalOpts, flags.Default)

func main() {
	epoch, e := strconv.ParseInt(buildStamp, 10, 64)
	if e == nil {
		buildDate = time.Unix(epoch, 0)
	}

	// TIME LIMITED BETA
	// 30 days self destruct
	if !buildDate.IsZero() && time.Since(buildDate) > 30*24*time.Hour {
		fatalln("This is an expired beta version.\nPlease grab a new build from http://ps-build1.vbg.se.prnw.net/job/mole")
	}
	// TIME LIMITED BETA
	// 30 days self destruct

	if runtime.GOOS == "windows" {
		globalOpts.Remap = true
	}

	globalParser.ApplicationName = "mole"
	if _, e := globalParser.Parse(); e != nil {
		if e, ok := e.(*flags.Error); ok {
			switch e.Type {
			case flags.ErrRequired:
				fmt.Println()
				globalParser.WriteHelp(os.Stdout)
				fmt.Println()
				fallthrough
			case flags.ErrHelp:
				fmt.Printf(msgExamples)
			}
		}
		os.Exit(1)
	}

	printTotalStats()
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

var setupDone bool

func setup() {
	if setupDone {
		return
	} else {
		setupDone = true
	}

	if globalOpts.NoAnsi {
		ansi.Disable()
	}

	if globalOpts.Debug {
		printVersion()
	}

	if strings.HasPrefix(globalOpts.Home, "~/") {
		home := getHomeDir()
		globalOpts.Home = strings.Replace(globalOpts.Home, "~", home, 1)
	}
	debugln("homeDir", globalOpts.Home)

	configFile := path.Join(globalOpts.Home, "mole.ini")
	f, e := os.Open(configFile)
	fatalErr(e)

	config := ini.Parse(f)
	serverIni.address = config.Sections["server"]["host"] + ":" + config.Sections["server"]["port"]
	serverIni.fingerprint = strings.ToLower(strings.Replace(config.Sections["server"]["fingerprint"], ":", "", -1))

	displayUpgradeNotice := true
	serverIni.upgrades = true
	if upgSec, ok := config.Sections["upgrades"]; ok {
		upgs, ok := upgSec["automatic"]
		displayUpgradeNotice = !ok
		serverIni.upgrades = !ok || upgs == "yes"
	}

	if serverIni.upgrades {
		go func() {
			time.Sleep(10 * time.Second)

			build, err := latestBuild()
			if err == nil {
				bd := time.Unix(int64(build.BuildStamp), 0)
				if isNewer := bd.Sub(buildDate).Seconds() > 0; isNewer {
					err = upgrade.UpgradeTo(build)
					if err == nil {
						if displayUpgradeNotice {
							infoln(msgAutoUpgrades)
						}
						okf(msgUpgraded, build.Version)
					}
				}
			}
		}()
	} else {
		debugln("automatic upgrades disabled")
	}
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
	cert, err := tls.LoadX509KeyPair(path.Join(globalOpts.Home, "mole.crt"), path.Join(globalOpts.Home, "mole.key"))
	fatalErr(err)
	return cert
}
