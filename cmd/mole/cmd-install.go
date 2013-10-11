// +build darwin linux

package main

import (
	"flag"
	"github.com/calmh/mole/table"
	"io"
	"io/ioutil"
	"os/exec"
	"runtime"
)

func init() {
	commands["install"] = command{installCommand, msgInstallShort}
}

func installCommand(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	fs.Usage = usageFor(fs, msgInstallUsage)
	fs.Parse(args)
	args = fs.Args()

	cl := NewClient(serverAddress(), moleIni.Get("server", "fingerprint"))
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
			infoln(table.FmtFunc("ll", rows, tableFormatter))
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

		_, err = io.Copy(wr, rd)
		fatalErr(err)
		err = wr.Close()
		fatalErr(err)
		_ = rd.Close()

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
}
