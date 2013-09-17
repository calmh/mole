package hosts

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type Entry struct {
	IP    string
	Names []string
	Tag   string
}

func ReplaceTagged(tag string, entries []Entry) error {
	hostf, err := os.Open("/etc/hosts")
	if err != nil {
		return err
	}

	var lines []string
	br := bufio.NewReader(hostf)
	for {
		bs, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		line := string(bs)
		if !strings.Contains(line, "#tag:"+tag) {
			lines = append(lines, line)
		}
	}
	hostf.Close()

	for _, e := range entries {
		lines = append(lines, e.IP+"\t"+strings.Join(e.Names, " ")+" #tag:"+tag)
	}

	outf, err := os.Create("/etc/hosts.new")
	if err != nil {
		return err
	}

	for _, line := range lines {
		_, err = outf.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	err = outf.Close()
	if err != nil {
		return err
	}

	return os.Rename("/etc/hosts.new", "/etc/hosts")
}
