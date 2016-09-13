// Package hosts manages alterations to the /etc/hosts file.
package hosts

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// An Entry represents a line in /etc/hosts
type Entry struct {
	IP    string
	Names []string
	Tag   string
}

// ReplaceTagged removes all lines in /etc/hosts marked with the specified tag
// and insert the given slice of entries instead. The slice may be nil or zero
// length to remove all tagged entries.
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
	_ = hostf.Close()

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
