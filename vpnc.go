package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/calmh/mole/configuration"
)

type Vpnc struct {
	cmd exec.Cmd
}

func vpnc(cfg *configuration.Config) *Vpnc {
	script := writeVpncScript(cfg)
	defer func() {
		debug("rm", script)
		os.Remove(script)
	}()

	// Absurd contortions with subshell because I can't seem to otherwise work
	// around "permission denied" when sending sigintr to the child.

	cmd := exec.Command("/bin/bash", "-c", "(sudo -p '(sudo) Account password, to invoke vpnc: ' /usr/local/sbin/vpnc --no-detach --non-inter --script "+script+" -)")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Started vpnc (pid %d)", cmd.Process.Pid)

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
			debug("vpn running")
			return &Vpnc{*cmd}
		}

		if err != nil {
			log.Fatal(err)
		}
	}
}

func (v *Vpnc) Stop() {
	log.Printf("Stopping vpnc (pid %d)", v.cmd.Process.Pid)
	e := v.cmd.Process.Signal(syscall.SIGTERM)
	if e != nil {
		log.Println(e)
	}
	v.cmd.Wait()
}

var lenToMaskMap = map[string]string{
	"0":  "0.0.0.0",
	"1":  "128.0.0.0",
	"2":  "192.0.0.0",
	"3":  "224.0.0.0",
	"4":  "240.0.0.0",
	"5":  "248.0.0.0",
	"6":  "252.0.0.0",
	"7":  "254.0.0.0",
	"8":  "255.0.0.0",
	"9":  "255.128.0.0",
	"10": "255.192.0.0",
	"11": "255.224.0.0",
	"12": "255.240.0.0",
	"13": "255.248.0.0",
	"14": "255.252.0.0",
	"15": "255.254.0.0",
	"16": "255.255.0.0",
	"17": "255.255.128.0",
	"18": "255.255.192.0",
	"19": "255.255.224.0",
	"20": "255.255.240.0",
	"21": "255.255.248.0",
	"22": "255.255.252.0",
	"23": "255.255.254.0",
	"24": "255.255.255.0",
	"25": "255.255.255.128",
	"26": "255.255.255.192",
	"27": "255.255.255.224",
	"28": "255.255.255.240",
	"29": "255.255.255.248",
	"30": "255.255.255.252",
	"31": "255.255.255.254",
	"32": "255.255.255.255",
}

func writeVpncScript(cfg *configuration.Config) string {
	f, e := ioutil.TempFile("", "vpnc-script.")
	if e != nil {
		log.Fatal(e)
	}
	debug(f.Name())

	_, e = f.Write([]byte(vpncScript(cfg)))
	if e != nil {
		log.Fatal(e)
	}

	e = f.Close()
	if e != nil {
		log.Fatal(e)
	}

	e = os.Chmod(f.Name(), 0x755)
	if e != nil {
		log.Fatal(e)
	}

	return f.Name()
}

func vpncScript(cfg *configuration.Config) string {
	script := `#!/bin/bash

has_init="no"
init() {
	if [ "$has_init" == "no" ] ; then
		has_init="yes"
		export CISCO_SPLIT_INC=0
		unset INTERNAL_IP4_DNS
	fi
}

add_route() {
	init
	export CISCO_SPLIT_INC_${CISCO_SPLIT_INC}_ADDR=$1
	export CISCO_SPLIT_INC_${CISCO_SPLIT_INC}_MASKLEN=$2
	export CISCO_SPLIT_INC_${CISCO_SPLIT_INC}_MASK=$3
	export CISCO_SPLIT_INC=$(($CISCO_SPLIT_INC + 1))
}

{add_cmds}

echo mole-vpnc-script-next
. /usr/local/etc/vpnc/vpnc-script
. /etc/vpnc/vpnc-script
`

	var addCmds []string
	for _, route := range cfg.VpnRoutes {
		ps := strings.SplitN(route, "/", 2)
		addCmds = append(addCmds, "add_route "+ps[0]+" "+ps[1]+" "+lenToMaskMap[ps[1]])
	}

	script = strings.Replace(script, "{add_cmds}", strings.Join(addCmds, "\n"), 1)
	debug(script)
	return script
}
