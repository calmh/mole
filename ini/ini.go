// Package ini provides trivial parsing of .INI format files.
package ini

import (
	"bufio"
	"io"
	"regexp"
	"strings"
)

// File is a parsed INI format file.
type File struct {
	Sections     map[string]Section
	SectionNames []string
}

// Section is a named [section] within a File.
type Section map[string]string

var (
	iniSectionRe = regexp.MustCompile(`^\[(.+)\]$`)
	iniOptionRe  = regexp.MustCompile(`^\s*([^\s]+)\s*=\s*(.+)\s*$`)
)

// Parse reads the given io.Reader and returns a parsed File object.
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

func (f *File) Write(out io.Writer) error {
	for _, sectionName := range f.SectionNames {
		_, err := out.Write([]byte("[" + sectionName + "]\n"))
		if err != nil {
			return err
		}
		for k, v := range f.Sections[sectionName] {
			_, err = out.Write([]byte(k + " = " + v + "\n"))
			if err != nil {
				return err
			}
		}
		_, err = out.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
