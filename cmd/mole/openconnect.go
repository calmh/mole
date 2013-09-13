// +build darwin linux

package main

import (
	"bufio"
	"log"
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
}

type OpenConnect struct {
	cmd    exec.Cmd
	script string
}

func (p OpenConnectProvider) Start(cfg *conf.Config) VPN {
	requireRoot("openconnect")

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
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if globalOpts.Debug {
		cmd.Stderr = os.Stderr
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf(msgOpncStart, cmd.Process.Pid)
	log.Println(msgOpncWait)

	stdin.Write([]byte(cfg.OpenConnect["password"] + "\n"))
	stdin.Close()

	bufReader := bufio.NewReader(stdout)
	for {
		bs, _, err := bufReader.ReadLine()
		line := strings.TrimSpace(string(bs))
		debug("opnc:", line)

		if strings.Contains(line, "Established DTLS connection") {
			log.Println(msgOpncConnected)
			log.Println()
			return &OpenConnect{*cmd, script}
		}

		if err != nil {
			log.Fatal(err)
		}
	}
}

func (v *OpenConnect) Stop() {
	defer func() {
		debug("rm", v.script)
		os.Remove(v.script)
	}()

	log.Printf(msgOpncStopping, v.cmd.Process.Pid)
	v.cmd.Process.Signal(os.Interrupt)
	v.cmd.Wait()
	log.Println(msgOpncStopped)
}
