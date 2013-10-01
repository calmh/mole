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
	sections []section
}

type section struct {
	name    string
	options []option
}

type option struct {
	name, value string
}

var (
	iniSectionRe = regexp.MustCompile(`^\[(.+)\]$`)
	iniOptionRe  = regexp.MustCompile(`^\s*([^\s]+)\s*=\s*(.+?)\s*$`)
)

// Sections returns the list of sections in the file.
func (f *File) Sections() []string {
	var sections []string
	for _, sect := range f.sections {
		sections = append(sections, sect.name)
	}
	return sections
}

// Options returns the list of options in a given section.
func (f *File) Options(section string) []string {
	var options []string
	for _, sect := range f.sections {
		if sect.name == section {
			for _, opt := range sect.options {
				options = append(options, opt.name)
			}
			break
		}
	}
	return options
}

// OptionMap returns the map option => value for a given section.
func (f *File) OptionMap(section string) map[string]string {
	options := make(map[string]string)
	for _, sect := range f.sections {
		if sect.name == section {
			for _, opt := range sect.options {
				options[opt.name] = opt.value
			}
			break
		}
	}
	return options
}

// Parse reads the given io.Reader and returns a parsed File object.
func Parse(stream io.Reader) File {
	var iniFile File
	var curSection section
	scanner := bufio.NewScanner(bufio.NewReader(stream))
	for scanner.Scan() {
		line := scanner.Text()
		if !(strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";")) && len(line) > 0 {
			if m := iniSectionRe.FindStringSubmatch(line); len(m) > 0 {
				if curSection.name != "" {
					iniFile.sections = append(iniFile.sections, curSection)
				}
				curSection = section{name: m[1]}
			} else if m := iniOptionRe.FindStringSubmatch(line); len(m) > 0 {
				val := strings.Trim(m[2], `"`)
				curSection.options = append(curSection.options, option{m[1], val})
			}
		}
	}
	if curSection.name != "" {
		iniFile.sections = append(iniFile.sections, curSection)
	}
	return iniFile
}

// Write writes the sections and options to the io.Writer in INI format.
func (f *File) Write(out io.Writer) error {
	for _, sect := range f.sections {
		out.Write([]byte("[" + sect.name + "]\n"))
		for _, opt := range sect.options {
			val := opt.value
			if len(val) == 0 || val[0] == ' ' || val[len(val)-1] == ' ' {
				val = `"` + val + `"`
			}
			out.Write([]byte(opt.name + " = " + val + "\n"))
		}
		out.Write([]byte("\n"))
	}
	return nil
}

// Get gets the value from the specified section and key name, or the empty
// string if either the section or the key is missing.
func (f *File) Get(section, key string) string {
	for _, sect := range f.sections {
		if sect.name == section {
			for _, opt := range sect.options {
				if opt.name == key {
					return opt.value
				}
			}
			return ""
		}
	}
	return ""
}

// Set sets a value for an option in a section. If the option exists, it's
// value will be overwritten. If the option does not exist, it will be added.
// If the section does not exist, it will be added and the option added to it.
func (f *File) Set(sectionName, key, value string) {
	for i, sect := range f.sections {
		if sect.name == sectionName {
			for j, opt := range sect.options {
				if opt.name == key {
					f.sections[i].options[j].value = value
					return
				}
			}
			f.sections[i].options = append(sect.options, option{key, value})
			return
		}
	}

	f.sections = append(f.sections, section{
		name:    sectionName,
		options: []option{{key, value}},
	})
}
