package main

import (
	"crypto/tls"
	"errors"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/calmh/mole/ini"
	"github.com/jessevdk/go-flags"
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

var globalParser = flags.NewParser(&globalOpts, flags.Default)

func main() {
	if runtime.GOOS == "windows" {
		globalOpts.Remap = true
	}

	globalParser.ApplicationName = "mole"
	if _, e := globalParser.Parse(); e != nil {
		os.Exit(1)
	}
}

func setup() {
	globalOpts.Debug = globalOpts.Debug || isDebug

	if globalOpts.Debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
		log.Println("Debug enabled")
	} else {
		log.SetFlags(0)
	}

	if globalOpts.Debug {
		printVersion()
	}

	userHome := os.Getenv("HOME")
	debug("userHome", userHome)
	if strings.HasPrefix(globalOpts.Home, "~/") {
		homeDir = strings.Replace(globalOpts.Home, "~", userHome, 1)
	}
	if globalOpts.Debug {
		log.Println("homeDir", homeDir)
	}

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
		log.Printf("  %s (%s)", buildVersion, buildKind)
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
