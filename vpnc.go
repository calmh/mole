package main

import (
	"bufio"
	"log"
	"os/exec"
	"strings"

	"github.com/calmh/mole/configuration"
)

func vpnc(cfg *configuration.Config) {
	var options []string
	options = append(options, "-p", "(sudo) Account password, to invoke \"vpnc\": ")
	options = append(options, "/usr/local/sbin/vpnc", "--no-detach", "--non-inter", "-")

	if globalOpts.Debug {
		log.Println(options)
	}

	var err error
	_ = bufio.MaxScanTokenSize
	cmd := exec.Command("sudo", options...)

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

	for k, v := range cfg.Vpnc {
		line := strings.Replace(k, "_", " ", -1) + " " + v + "\n"
		stdin.Write([]byte(line))
	}
	stdin.Close()

	buf := make([]byte, 1024)
	for {
		n, err := stdout.Read(buf)
		s := strings.TrimSpace(string(buf[:n]))
		log.Println(s)
		if err != nil {
			log.Fatal(err)
		}
	}

	cmd.Wait()
}
