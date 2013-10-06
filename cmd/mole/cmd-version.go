package main

import (
	"runtime"
	"sync"
	"time"
)

func init() {
	commands["version"] = command{versionCommand, msgVersionShort}
}

func versionCommand(args []string) error {
	printVersion()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))
		if ver := cl.ServerVersion(); ver != "" {
			infof("Server version:\n  %s", ver)
		}
		wg.Done()
	}()

	go func() {
		build, err := latestBuild()
		if err == nil {
			infof("Latest client version:\n  %s", build.Version)

			bd := time.Unix(int64(build.BuildStamp), 0)
			if isNewer := bd.Sub(buildDate).Seconds() > 0; isNewer {
				warnln("Run 'mole upgrade'.")
			} else {
				okln(msgLatest)
			}
		}
		wg.Done()
	}()

	wg.Wait()

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
