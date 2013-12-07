package main

import (
	"fmt"
	"log"
	"os/exec"
)

func gitCommit(dir, comment, user string) {
	var cmd *exec.Cmd

	cmd = exec.Command("git", "add", "-A")
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("git:", err)
		return
	}

	author := fmt.Sprintf("%s <%s@mole>", user, user)
	cmd = exec.Command("git", "commit", "--author", author, "-m", comment)
	cmd.Dir = dir
	_, err = cmd.CombinedOutput()
	if err != nil {
		log.Println("git:", err)
		return
	}
}

func gitInit(dir string) {
	var cmd *exec.Cmd

	cmd = exec.Command("git", "init")
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("git:", err)
		return
	}
}
