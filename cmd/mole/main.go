package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
	"nym.se/mole/ini"
)

var errParams = errors.New("incorrect command line parameters")

var (
	buildVersion string
	buildDate    string
	buildUser    string
	homeDir      string
	serverAddr   string
)

var globalOpts struct {
	Debug bool   `short:"d" long:"debug" description:"Show debug output"`
	Home  string `short:"h" long:"home" description:"Mole home directory" default:"~/.mole" value-name:"DIR"`
	Remap bool   `long:"remap-lo" description:"Use port remapping for extended lo addresses (required/default on Windows)"`
}

var globalStats struct {
	dataIn  uint64
	dataOut uint64
}

var globalParser = flags.NewParser(&globalOpts, flags.Default)

func main() {
	dropRoot()

	if runtime.GOOS == "windows" {
		globalOpts.Remap = true
	}

	globalParser.ApplicationName = "mole"
	if _, e := globalParser.Parse(); e != nil {
		os.Exit(1)
	}

	log.Println(msgOk)
	if globalStats.dataIn+globalStats.dataOut > 0 {
		printStatistics()
	}
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

func printStatistics() {
	log.Printf(" -- %sB in, %sB out", formatBytes(globalStats.dataIn), formatBytes(globalStats.dataOut))
}

func setup() {
	if globalOpts.Debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
		log.Println(msgDebugEnabled)
	} else {
		log.SetFlags(0)
	}

	if globalOpts.Debug {
		printVersion()
	}

	userHome := os.Getenv("HOME")
	debug("userHome", userHome)
	if userHome == "" {
		log.Fatal(msgErrNoHome)
	}

	if strings.HasPrefix(globalOpts.Home, "~/") {
		homeDir = strings.Replace(globalOpts.Home, "~", userHome, 1)
	}
	debug("homeDir", homeDir)

	configFile := path.Join(homeDir, "mole.ini")
	f, e := os.Open(configFile)
	if e != nil {
		log.Fatal(e)
	}

	config := ini.Parse(f)
	serverAddr = config.Sections["server"]["host"] + ":" + config.Sections["server"]["port"]
}

func printVersion() {
	log.Printf("mole (%s-%s)", runtime.GOOS, runtime.GOARCH)
	if buildVersion != "" {
		log.Printf("  %s", buildVersion)
	}
	if buildDate != "" {
		log.Printf("  %s by %s", buildDate, buildUser)
	}
}

func certificate() tls.Certificate {
	cert, err := tls.LoadX509KeyPair(path.Join(homeDir, "mole.crt"), path.Join(homeDir, "mole.key"))
	if err != nil {
		log.Fatal(err)
	}
	return cert
}
