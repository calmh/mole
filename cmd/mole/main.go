package main

import (
	"flag"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/calmh/ini"
	"github.com/calmh/mole/ansi"
	"github.com/calmh/mole/upgrade"
)

var (
	buildVersion string
	buildStamp   string
	buildDate    time.Time
	buildUser    string
	homeDir      string = path.Join(getHomeDir(), ".mole")
	debugEnabled bool
	useAnsi      bool = isTerminal(os.Stdout.Fd())
	remapIntfs   bool
)

var moleIni ini.Config

type command struct {
	name    string
	fn      func([]string)
	descr   string
	aliases []string
}

var commandList []command
var commandMap = make(map[string]command)

func addCommand(c command) {
	commandList = append(commandList, c)
	commandMap[c.name] = c
	for _, alias := range c.aliases {
		commandMap[alias] = c
	}
}

func init() {
	epoch, e := strconv.ParseInt(buildStamp, 10, 64)
	if e == nil {
		buildDate = time.Unix(epoch, 0)
	}
}

func main() {
	loadConfig()

	// Early disable ansi, if the defaults say to do so.
	if !useAnsi {
		ansi.Disable()
	}

	// Using the logging functions before this point will panic. Used to weed
	// out logging before having set up the debug & ansi flags.
	enableLogging()

	// Ensure that we have a home directory and that it has the right permissions.
	ensureHome()

	// Set the default remapIntfs before parsing flags
	if runtime.GOOS == "windows" {
		remapIntfs = true
	}

	args := parseFlags()

	// Late disable ansi, if the command line flags said so.
	if !useAnsi {
		ansi.Disable()
	}

	if debugEnabled {
		printVersion()
	}

	go autoUpgrade()

	// Early beta versions of mole4 wrote the fingerprint in lower case which
	// is incompatible with both mole 3 and current 4+. Rewrite the fingerprint
	// to upper case if necessary.
	if fp := moleIni.Get("server", "fingerprint"); fp != strings.ToUpper(fp) {
		moleIni.Set("server", "fingerprint", strings.ToUpper(fp))
		saveMoleIni()
	}

	dispatchCommand(args)

	exit(0)
}

// Keep this short and sweet so we get can call it very early and get default
// values for the debug and ansi flags.
func loadConfig() {
	configFile := path.Join(homeDir, "mole.ini")
	if fd, err := os.Open(configFile); err == nil {
		moleIni = ini.Parse(fd)
	}

	debugEnabled = moleIni.Get("client", "debug") == "yes"
}

func dispatchCommand(args []string) {
	// Direct match on command
	if cmd, ok := commandMap[args[0]]; ok {
		cmd.fn(args[1:])
		exit(0)
	}

	// Unique prefix match
	var found string
	for n := range commandMap {
		if strings.HasPrefix(n, args[0]) {
			if found != "" && commandMap[found].name != commandMap[n].name {
				fatalf("ambigous command: %q (could be %q or %q)", args[0], n, found)
			}
			found = n
		}
	}
	if found != "" {
		cmd := commandMap[found]
		cmd.fn(args[1:])
		exit(0)
	}

	// No command found
	fatalf("no such command: %q", args[0])
}

// Ensure home direcory exists and has appropriate permissions.
func ensureHome() {
	fi, err := os.Stat(homeDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(homeDir, 0700)
		fatalErr(err)
		okln("Created", homeDir)
	} else if fi.Mode()&0077 != 0 {
		err := os.Chmod(homeDir, 0700)
		fatalErr(err)
		okln("Corrected permissions on", homeDir)
	}
}

func parseFlags() []string {
	fs := flag.NewFlagSet("mole", flag.ContinueOnError)
	fs.BoolVar(&debugEnabled, "d", debugEnabled, "Enable debug output")
	fs.BoolVar(&useAnsi, "ansi", useAnsi, "Enable/disable ANSI formatting")
	fs.BoolVar(&remapIntfs, "remap", remapIntfs, "Use port remapping for extended lo addresses")
	fs.Usage = usageFor(fs, msgMainUsage)
	err := fs.Parse(os.Args[1:])

	if err != nil {
		// fs.Usage() has already been printed
		mainUsage(os.Stdout)
		exit(3)
	}

	args := fs.Args()
	if len(args) == 0 {
		fs.Usage()
		mainUsage(os.Stdout)
		exit(3)
	}

	return args
}

func autoUpgrade() {
	if moleIni.Get("upgrades", "automatic") == "no" {
		debugln("automatic upgrades disabled")
		return
	}

	// Only do the actual upgrade once we've been running for a while
	time.Sleep(10 * time.Second)
	build, err := latestBuild()
	if err == nil {
		bd := time.Unix(int64(build.BuildStamp), 0)
		if isNewer := bd.Sub(buildDate).Seconds() > 0; isNewer {
			err = upgrade.UpgradeTo(build)
			if err == nil {
				if moleIni.Get("upgrades", "automatic") != "yes" {
					infoln(msgAutoUpgrades)
				}
				okf(msgUpgraded, build.Version)
			} else {
				warnln("Automatic upgrade failed:", err)
			}
		}
	}
}

func serverAddress() string {
	port, _ := strconv.Atoi(moleIni.Get("server", "port"))
	return moleIni.Get("server", "host") + ":" + strconv.Itoa(port)
}

func saveMoleIni() {
	fd, err := os.Create(path.Join(homeDir, "mole.ini"))
	fatalErr(err)
	err = moleIni.Write(fd)
	fatalErr(err)
	err = fd.Close()
	fatalErr(err)
}
