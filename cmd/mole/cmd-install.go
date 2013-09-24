// +build darwin linux

package main

import (
	"github.com/calmh/mole/table"
	"github.com/jessevdk/go-flags"
	"io"
	"io/ioutil"
	"os/exec"
	"runtime"
)

type cmdInstall struct{}

var installParser *flags.Parser

func init() {
	cmd := cmdInstall{}
	installParser = globalParser.AddCommand("install", msgInstallShort, msgInstallLong, &cmd)
}

func (c *cmdInstall) Usage() string {
	return "[package] [install-OPTIONS]"
}

func (c *cmdInstall) Execute(args []string) error {
	setup()

	cl := NewClient(serverIni.address, serverIni.fingerprint)
	if len(args) == 0 {
		pkgMap, err := cl.Packages()
		fatalErr(err)

		arch := runtime.GOOS + "-" + runtime.GOARCH
		var rows [][]string
		for _, pkg := range pkgMap[arch] {
			rows = append(rows, []string{pkg.Package, pkg.Description})
		}

		if len(rows) > 0 {
			rows = append([][]string{{"PKG", "DESCRIPTION"}}, rows...)
			infoln(table.Fmt("ll", rows))
		} else {
			infoln(msgNoPackages)
		}
	} else {
		requireRoot("install")
		name := args[0]
		fullname := args[0] + "-" + runtime.GOOS + "-" + runtime.GOARCH + ".tar.gz"
		wr, err := ioutil.TempFile("", name)
		fatalErr(err)

		rd, err := cl.Package(fullname)
		fatalErr(err)

		io.Copy(wr, rd)
		wr.Close()
		rd.Close()

		td, err := ioutil.TempDir("", name)
		fatalErr(err)

		cmd := exec.Command("sh", "-c", "cd "+td+" && tar zxf "+wr.Name())
		_, err = cmd.CombinedOutput()
		fatalErr(err)

		becomeRoot()
		cmd = exec.Command(td+"/install.sh", td)
		_, err = cmd.CombinedOutput()
		fatalErr(err)
		dropRoot()

		okln("Installed")
	}

	return nil
}
