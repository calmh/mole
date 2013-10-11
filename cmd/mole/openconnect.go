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
}

type OpenConnect struct {
	cmd    exec.Cmd
	script string
}

func (p OpenConnectProvider) Start(cfg *conf.Config) (VPN, error) {
	requireRoot("openconnect")

	if ok := ensureTunModule(); !ok {
		return nil, fmt.Errorf(msgErrNoTunModule)
	}

	script := writeVpncScript(cfg)

	args := []string{"--non-inter", "--passwd-on-stdin", "--script", script}
	for _, kv := range cfg.OpenConnect {
		if kv.Key == "password" {
			continue
		} else if kv.Key == "server" {
			args = append(args, kv.Value)
		} else if kv.Value == "yes" {
			args = append(args, "--"+kv.Key)
		} else {
			args = append(args, "--"+kv.Key+"="+kv.Value)
		}

	}
	cmd := exec.Command(p.openconnectBinary, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if debugEnabled {
		cmd.Stderr = os.Stderr
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	debugf(msgOpncStart, cmd.Process.Pid)
	debugln(msgOpncWait)

	_, err = stdin.Write([]byte(cfg.OpenConnect.Get("password") + "\n"))
	if err != nil {
		return nil, err
	}
	err = stdin.Close()
	if err != nil {
		return nil, err
	}

	bufReader := bufio.NewReader(stdout)
	for {
		bs, _, err := bufReader.ReadLine()
		line := strings.TrimSpace(string(bs))
		debugln("opnc:", line)

		if strings.Contains(line, "Established DTLS connection") {
			debugln(msgOpncConnected)
			return &OpenConnect{*cmd, script}, nil
		}

		if err != nil {
			return nil, err
		}
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

	debugf(msgOpncStopping, v.cmd.Process.Pid)
	err := v.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		warnln(err)
	}
	err = v.cmd.Wait()
	if err != nil {
		warnln(err)
	}
	debugln(msgOpncStopped)
}
