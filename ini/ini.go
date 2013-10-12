// Package ini provides trivial parsing of .INI format files.
package ini

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// File is a parsed INI format file.
type File struct {
	sections []section
}

type section struct {
	name     string
	comments []string
	options  []option
}

type option struct {
	name, value string
}

var (
	iniSectionRe = regexp.MustCompile(`^\[(.+)\]$`)
	iniOptionRe  = regexp.MustCompile(`^([^\s]+)\s*=\s*(.+?)$`)
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

// Comments returns the list of comments in a given section.
func (f *File) Comments(section string) []string {
	for _, sect := range f.sections {
		if sect.name == section {
			return sect.comments
		}
	}
	return nil
}

// Parse reads the given io.Reader and returns a parsed File object.
func Parse(stream io.Reader) File {
	var iniFile File
	var curSection section
	scanner := bufio.NewScanner(bufio.NewReader(stream))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			comment := strings.TrimLeft(line, ";# ")
			curSection.comments = append(curSection.comments, comment)
		} else if len(line) > 0 {
			if m := iniSectionRe.FindStringSubmatch(line); len(m) > 0 {
				if curSection.name != "" {
					iniFile.sections = append(iniFile.sections, curSection)
				}
				curSection = section{name: m[1]}
			} else if m := iniOptionRe.FindStringSubmatch(line); len(m) > 0 {
				val := strings.Trim(m[2], `"`)
				val = strings.Replace(val, "\\n", "\n", -1)
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
		fmt.Fprintf(out, "[%s]\n", sect.name)
		for _, cmt := range sect.comments {
			fmt.Fprintln(out, "; "+cmt)
		}
		for _, opt := range sect.options {
			val := opt.value
			if len(val) == 0 {
				continue
			}
			if val[0] == ' ' || val[len(val)-1] == ' ' || strings.Contains(val, "\n") {
				val = fmt.Sprintf("%q", val)
			}
			fmt.Fprintf(out, "%s=%s\n", opt.name, val)
		}
		fmt.Fprintln(out)
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
