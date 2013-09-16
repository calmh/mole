package main

import (
	"github.com/jessevdk/go-flags"
	"nym.se/mole/upgrade"
	"time"
)

type cmdUpgrade struct {
	Silently bool `long:"silent" description:"Say nothing unless an upgrade was performed"`
	Force    bool `long:"force" description:"Don't perform newness check, just upgrade to whatever the server has."`
}

var upgradeParser *flags.Parser

func init() {
	cmd := cmdUpgrade{}
	upgradeParser = globalParser.AddCommand("upgrade", "Upgrade mole", "'upgrade' checks for the latest version of the mole client and performs upgrades", &cmd)
}

func (c *cmdUpgrade) Execute(args []string) error {
	setup()

	cert := certificate()
	cl := NewClient(serverAddr, cert)

	upgradesURL := cl.UpgradesURL()
	if upgradesURL == "" {
		if c.Silently {
			debugln(msgNoUpgradeURL)
		} else {
			warnln(msgNoUpgradeURL)
		}
		return nil
	}

	if c.Silently {
		debugf(msgCheckingUpgrade, upgradesURL)
	} else {
		infof(msgCheckingUpgrade, upgradesURL)
	}

	build, err := upgrade.Newest("mole", upgradesURL)
	if err == nil {
		bd := time.Unix(int64(build.BuildStamp), 0)
		isNewer := bd.Sub(buildDate).Seconds() > 0
		if c.Force || isNewer {
			if c.Silently {
				debugln(msgDownloadingUpgrade)
			} else {
				infoln(msgDownloadingUpgrade)
			}

			err = upgrade.UpgradeTo(build)
			if err != nil {
				warnln(err)
			} else {
				okf(msgUpgraded, build.Version)
			}
		} else {
			if c.Silently {
				debugln(msgNoUpgrades)
			} else {
				okln(msgNoUpgrades)
			}
		}
	} else {
		debugln(err)
	}

	return nil
}
