// +build darwin linux

package main

import (
	"bufio"
	"fmt"
	"github.com/calmh/mole/conf"
	"os"
	"os/exec"
	"strings"
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

func (p VPNCProvider) Start(cfg *conf.Config) (VPN, error) {
	requireRoot("vpnc")

	if ok := ensureTunModule(); !ok {
		return nil, fmt.Errorf(msgErrNoTunModule)
	}

	script := writeVpncScript(cfg)
	cmd := exec.Command(p.vpncBinary, "--no-detach", "--non-inter", "--script", script, "-")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if globalOpts.Debug {
		cmd.Stderr = os.Stderr
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	debugf(msgVpncStart, cmd.Process.Pid)
	debugln(msgVpncWait)

	for k, v := range cfg.Vpnc {
		line := strings.Replace(k, "_", " ", -1) + " " + v + "\n"
		_, err := stdin.Write([]byte(line))
		if err != nil {
			return nil, err
		}
	}
	err = stdin.Close()
	if err != nil {
		return nil, err
	}

	bufReader := bufio.NewReader(stdout)
	for {
		bs, _, err := bufReader.ReadLine()
		line := strings.TrimSpace(string(bs))
		debugln("vpnc:", line)

		if line == "mole-vpnc-script-next" {
			debugln(msgVpncConnected)
			return &Vpnc{*cmd, script}, nil
		}

		if err != nil {
			return nil, err
		}
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

	debugln(msgVpncStopping, v.cmd.Process.Pid)
	err := v.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		warnln(err)
	}
	err = v.cmd.Wait()
	if err != nil {
		warnln(err)
	}
	debugln(msgVpncStopped)
}
