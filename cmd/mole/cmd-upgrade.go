package main

import (
	"errors"
	"flag"
	"time"

	"github.com/calmh/mole/upgrade"
	"github.com/calmh/mole/usage"
)

var errNoUpgradeUrl = errors.New("no upgrade URL")

func init() {
	addCommand(command{name: "upgrade", fn: upgradeCommand, descr: msgUpgradeShort})
}

func latestBuild() (build upgrade.Build, err error) {
	cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))

	upgradesURL, err := cl.UpgradesURL()
	if err != nil {
		return
	}
	if upgradesURL == "" {
		err = errNoUpgradeUrl
		return
	}

	debugln("checking for upgrade at", upgradesURL)
	build, err = upgrade.Newest("mole", upgradesURL)
	debugln("got build", build)
	return
}

func upgradeCommand(args []string) {
	fs := flag.NewFlagSet("upgrade", flag.ExitOnError)
	force := fs.Bool("force", false, "Perform upgrade to same or older version")
	disableAuto := fs.Bool("disable-auto", false, "Disable automatic upgrades")
	enableAuto := fs.Bool("enable-auto", false, "Enable automatic upgrades")
	fs.Usage = usage.For(fs, msgUpgradeUsage)
	fs.Parse(args)
	args = fs.Args()

	if *disableAuto || *enableAuto {
		upgrades := "yes"
		if *disableAuto {
			upgrades = "no"
		}
		moleIni.Set("upgrades", "automatic", upgrades)
		saveMoleIni()
		okf("Automatic upgrades set to %q", upgrades)
		return
	}

	build, err := latestBuild()
	fatalErr(err)

	bd := time.Unix(int64(build.BuildStamp), 0)
	isNewer := bd.Sub(buildDate).Seconds() > 0
	if *force || isNewer {
		infoln(msgDownloadingUpgrade)

		err = upgrade.UpgradeTo(build)
		fatalErr(err)
		okf(msgUpgraded, build.Version)
	} else {
		okln(msgLatest)
	}
}
