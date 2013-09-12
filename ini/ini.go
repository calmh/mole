package ini

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

type Section map[string]string
type File struct {
	Sections     map[string]Section
	SectionNames []string
}

var (
	iniSectionRe = regexp.MustCompile(`^\[(.+)\]$`)
	iniOptionRe  = regexp.MustCompile(`^\s*([^\s]+)\s*=\s*(.+)\s*$`)
)

func Parse(stream io.Reader) File {
	iniFile := File{Sections: make(map[string]Section)}
	var curSection string
	scanner := bufio.NewScanner(bufio.NewReader(stream))
	for scanner.Scan() {
		line := scanner.Text()
		if !(strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";")) && len(line) > 0 {
			if m := iniSectionRe.FindStringSubmatch(line); len(m) > 0 {
				curSection = m[1]
				iniFile.Sections[curSection] = make(Section)
				iniFile.SectionNames = append(iniFile.SectionNames, curSection)
			} else if m := iniOptionRe.FindStringSubmatch(line); len(m) > 0 {
				iniFile.Sections[curSection][m[1]] = strings.Trim(m[2], `"`)
			}
		}
	}
	return iniFile
}
