// +build darwin linux

package main

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"nym.se/mole/conf"
)

type OpenConnectProvider struct {
	openconnectBinary string
}

func init() {
	locations := []string{
		"/usr/bin/openconnect",
		"/usr/sbin/openconnect",
		"/usr/local/bin/openconnect",
		"/usr/local/sbin/openconnect",
	}
	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			vpnProviders["openconnect"] = OpenConnectProvider{path}
			return
		}
	}
	warnln(msgOpncUnavailable)
}

type OpenConnect struct {
	cmd    exec.Cmd
	script string
}

func (p OpenConnectProvider) Start(cfg *conf.Config) VPN {
	requireRoot("openconnect")

	if ok := ensureTunModule(); !ok {
		fatalln(msgErrNoTunModule)
	}

	script := writeVpncScript(cfg)

	args := []string{"--non-inter", "--passwd-on-stdin", "--script", script}
	for k, v := range cfg.OpenConnect {
		if k == "password" {
			continue
		} else if k == "server" {
			args = append(args, v)
		} else if v == "yes" {
			args = append(args, "--"+k)
		} else {
			args = append(args, "--"+k+"="+v)
		}

	}
	cmd := exec.Command(p.openconnectBinary, args...)

	stdin, err := cmd.StdinPipe()
	fatalErr(err)

	stdout, err := cmd.StdoutPipe()
	fatalErr(err)

	if globalOpts.Debug {
		cmd.Stderr = os.Stderr
	}

	err = cmd.Start()
	fatalErr(err)
	infof(msgOpncStart, cmd.Process.Pid)
	infoln(msgOpncWait)

	_, err = stdin.Write([]byte(cfg.OpenConnect["password"] + "\n"))
	fatalErr(err)
	err = stdin.Close()
	fatalErr(err)

	bufReader := bufio.NewReader(stdout)
	for {
		bs, _, err := bufReader.ReadLine()
		line := strings.TrimSpace(string(bs))
		debugln("opnc:", line)

		if strings.Contains(line, "Established DTLS connection") {
			infoln(msgOpncConnected)
			return &OpenConnect{*cmd, script}
		}

		fatalErr(err)
	}
}

func (v *OpenConnect) Stop() {
	defer func() {
		debugln("rm", v.script)
		err := os.Remove(v.script)
		if err != nil {
			warnln(err)
		}
	}()

	infof(msgOpncStopping, v.cmd.Process.Pid)
	err := v.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		warnln(err)
	}
	err = v.cmd.Wait()
	if err != nil {
		warnln(err)
	}
	infoln(msgOpncStopped)
}
