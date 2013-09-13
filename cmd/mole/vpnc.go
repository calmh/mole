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
}

type Vpnc struct {
	cmd    exec.Cmd
	script string
}

func (p VPNCProvider) Start(cfg *conf.Config) VPN {
	requireRoot("vpnc")

	script := writeVpncScript(cfg)
	cmd := exec.Command(p.vpncBinary, "--no-detach", "--non-inter", "--script", script, "-")

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
	log.Printf(msgVpncStart, cmd.Process.Pid)
	log.Println(msgVpncWait)

	for k, v := range cfg.Vpnc {
		line := strings.Replace(k, "_", " ", -1) + " " + v + "\n"
		stdin.Write([]byte(line))
	}
	stdin.Close()

	bufReader := bufio.NewReader(stdout)
	for {
		bs, _, err := bufReader.ReadLine()
		line := strings.TrimSpace(string(bs))
		debug("vpnc:", line)

		if line == "mole-vpnc-script-next" {
			log.Println(msgVpncConnected)
			log.Println()
			return &Vpnc{*cmd, script}
		}

		if err != nil {
			log.Fatal(err)
		}
	}
}

func (v *Vpnc) Stop() {
	defer func() {
		debug("rm", v.script)
		os.Remove(v.script)
	}()

	log.Printf(msgVpncStopping, v.cmd.Process.Pid)
	v.cmd.Process.Signal(os.Interrupt)
	v.cmd.Wait()
	log.Println(msgVpncStopped)
}
