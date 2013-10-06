package main

import (
	"runtime"
	"time"
)

func init() {
	commands["version"] = command{versionCommand, msgVersionShort}
}

func versionCommand(args []string) error {
	printVersion()

	cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))
	if ver := cl.ServerVersion(); ver != "" {
		infof("Server version:")
		infof("  %s", ver)
	}

	build, err := latestBuild()
	if err == nil {
		infoln("Latest client version:")
		infof("  %s", build.Version)

		bd := time.Unix(int64(build.BuildStamp), 0)
		if isNewer := bd.Sub(buildDate).Seconds() > 0; isNewer {
			warnln("Run 'mole upgrade'.")
		} else {
			okln(msgLatest)
		}
	}

	return nil
}

func printVersion() {
	infof("Client version (mole %s-%s):", runtime.GOOS, runtime.GOARCH)
	if buildVersion != "" {
		infof("  %s", buildVersion)
	}
	if !buildDate.IsZero() {
		infof("  %v by %s", buildDate, buildUser)
	}
}
