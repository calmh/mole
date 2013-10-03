package main

import (
	"flag"
	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/ini"
	"github.com/calmh/mole/upgrade"
	"io"
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
	Home       string
	Debug      bool
	NoAnsi     bool
	Remap      bool
	PortOffset int
}

var serverIni struct {
	address       string
	upgrades      bool
	fingerprint   string
	ticket        string
	upgradeNotice bool
}

type command struct {
	fn    func([]string) error
	descr string
}

var commands = make(map[string]command)

func main() {
	epoch, e := strconv.ParseInt(buildStamp, 10, 64)
	if e == nil {
		buildDate = time.Unix(epoch, 0)
	}

	if runtime.GOOS == "windows" {
		globalOpts.Remap = true
	}

	fs := flag.NewFlagSet("mole", flag.ContinueOnError)
	fs.StringVar(&globalOpts.Home, "home", "~/.mole", "Set mole's home directory")
	fs.BoolVar(&globalOpts.Debug, "d", false, "Enable debug output")
	fs.BoolVar(&globalOpts.NoAnsi, "no-ansi", false, "Disable ANSI formatting")
	fs.BoolVar(&globalOpts.Remap, "remap", globalOpts.Remap, "Use port remapping for extended lo addresses")
	fs.IntVar(&globalOpts.PortOffset, "port-offset", 1000, "**Temp** v3/v4 server compatibility (port shift)")
	fs.Usage = usageFor(fs, msgMainUsage)
	err := fs.Parse(os.Args[1:])

	if err != nil {
		// fs.Usage() has already been printed
		mainUsage(os.Stdout)
		os.Exit(3)
	}

	setup()

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		mainUsage(os.Stdout)
		os.Exit(3)
	}

	// Direct match on command
	if cmd, ok := commands[args[0]]; ok {
		cmd.fn(args[1:])
		os.Exit(0)
	}

	// Unique prefix match
	var found string
	for n := range commands {
		if strings.HasPrefix(n, args[0]) {
			if found != "" {
				fatalf("ambigous command: %q (could be %q or %q)", args[0], n, found)
			}
			found = n
		}
	}
	if found != "" {
		cmd := commands[found]
		cmd.fn(args[1:])
		os.Exit(0)
	}

	// No command found
	fatalf("no such command: %q", args[0])
}

func setup() {
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

	if fd, err := os.Open(configFile); err == nil {
		loadGlobalIni(fd)
		if serverIni.upgrades {
			go autoUpgrade()
		} else {
			debugln("automatic upgrades disabled")
		}
	}

	fi, err := os.Stat(globalOpts.Home)
	if os.IsNotExist(err) {
		err := os.MkdirAll(globalOpts.Home, 0700)
		fatalErr(err)
		okln("Created", globalOpts.Home)
	} else if fi.Mode()&0077 != 0 {
		err := os.Chmod(globalOpts.Home, 0700)
		fatalErr(err)
		okln("Corrected permissions on", globalOpts.Home)
	}
}

func loadGlobalIni(fd io.Reader) {
	config := ini.Parse(fd)
	port, _ := strconv.Atoi(config.Get("server", "port"))
	port += globalOpts.PortOffset
	serverIni.address = config.Get("server", "host") + ":" + strconv.Itoa(port)
	serverIni.fingerprint = config.Get("server", "fingerprint")
	serverIni.ticket = config.Get("server", "ticket")
	serverIni.upgrades = config.Get("upgrades", "automatic") != "no"
	serverIni.upgradeNotice = config.Get("upgrades", "automatic") != "yes"
}

func autoUpgrade() {
	// Only do the actual upgrade once we've been running for a while
	time.Sleep(10 * time.Second)
	build, err := latestBuild()
	if err == nil {
		bd := time.Unix(int64(build.BuildStamp), 0)
		if isNewer := bd.Sub(buildDate).Seconds() > 0; isNewer {
			err = upgrade.UpgradeTo(build)
			if err == nil {
				if serverIni.upgradeNotice {
					infoln(msgAutoUpgrades)
				}
				okf(msgUpgraded, build.Version)
			}
		}
	}
}
