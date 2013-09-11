package ini

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

type Section map[string]string
type File map[string]Section

var (
	iniSectionRe = regexp.MustCompile(`^\[(.+)\]$`)
	iniOptionRe  = regexp.MustCompile(`^\s*([^\s]+)\s*=\s*(.+)\s*$`)
)

func Parse(stream io.Reader) File {
	iniFile := make(File)
	var curSection string
	scanner := bufio.NewScanner(bufio.NewReader(stream))
	for scanner.Scan() {
		line := scanner.Text()
		if !(strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";")) && len(line) > 0 {
			if m := iniSectionRe.FindStringSubmatch(line); len(m) > 0 {
				curSection = m[1]
				iniFile[curSection] = make(Section)
			} else if m := iniOptionRe.FindStringSubmatch(line); len(m) > 0 {
				iniFile[curSection][m[1]] = strings.Trim(m[2], `"`)
			}
		}
	}
	return iniFile
}
