// +build darwin linux

package main

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"nym.se/mole/conf"
)

type VPNCProvider struct {
	vpncBinary string
}

func init() {
	locations := []string{
		"/usr/bin/vpnc",
		"/usr/sbin/vpnc",
		"/usr/local/bin/vpnc",
		"/usr/local/sbin/vpnc",
	}
	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			vpnProviders["vpnc"] = VPNCProvider{path}
			return
		}
	}
	warnln(msgVpncUnavailable)
}

type Vpnc struct {
	cmd    exec.Cmd
	script string
}

func (p VPNCProvider) Start(cfg *conf.Config) VPN {
	requireRoot("vpnc")

	if ok := ensureTunModule(); !ok {
		fatalln(msgErrNoTunModule)
	}

	script := writeVpncScript(cfg)
	cmd := exec.Command(p.vpncBinary, "--no-detach", "--non-inter", "--script", script, "-")

	stdin, err := cmd.StdinPipe()
	fatalErr(err)

	stdout, err := cmd.StdoutPipe()
	fatalErr(err)

	if globalOpts.Debug {
		cmd.Stderr = os.Stderr
	}

	err = cmd.Start()
	fatalErr(err)
	infof(msgVpncStart, cmd.Process.Pid)
	infoln(msgVpncWait)

	for k, v := range cfg.Vpnc {
		line := strings.Replace(k, "_", " ", -1) + " " + v + "\n"
		_, err := stdin.Write([]byte(line))
		fatalErr(err)
	}
	err = stdin.Close()
	fatalErr(err)

	bufReader := bufio.NewReader(stdout)
	for {
		bs, _, err := bufReader.ReadLine()
		line := strings.TrimSpace(string(bs))
		debugln("vpnc:", line)

		if line == "mole-vpnc-script-next" {
			infoln(msgVpncConnected)
			return &Vpnc{*cmd, script}
		}

		fatalErr(err)
	}
}

func (v *Vpnc) Stop() {
	defer func() {
		debugln("rm", v.script)
		err := os.Remove(v.script)
		if err != nil {
			warnln(err)
		}
	}()

	infof(msgVpncStopping, v.cmd.Process.Pid)
	err := v.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		warnln(err)
	}
	err = v.cmd.Wait()
	if err != nil {
		warnln(err)
	}
	infoln(msgVpncStopped)
}
