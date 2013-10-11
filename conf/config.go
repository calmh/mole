package conf

import (
	"encoding/xml"
	"fmt"
	"github.com/calmh/mole/ini"
	"io"
	"sort"
)

const (
	defaultSSHPort = 22
)

// Features that are required to correctly handle a tunnel configuration
const (
	FeatureError uint32 = 1 << iota
	FeatureSshPassword
	FeatureSshKey
	FeatureVpnc
	FeatureOpenConnect
	FeatureLocalOnly
	FeatureSocks
)

// Config is a complete tunnel configuration
type Config struct {
	General struct {
		Description  string       `xml:"description"`
		AuthorName   string       `xml:"author>name"`
		AuthorEmail  string       `xml:"author>email"`
		Main         string       `xml:"main,omitempty"`
		Version      int          `xml:"-"`
		VersionFloat float64      `xml:"version,attr"`
		Other        KeyValueList `xml:"other,omitempty"`
	} `xml:"meta"`

	Hosts       []Host       `xml:"host"`
	Forwards    []Forward    `xml:"forward"`
	OpenConnect KeyValueList `xml:"openconnect>opt,omitempty"`
	Vpnc        KeyValueList `xml:"vpnc>opt,omitempty"`
	VpnRoutes   []string     `xml:"vpnroutes>prefix,omitempty"`

	XMLName struct{} `xml:"tunnel"`
}

// Host is an SSH host to bounce via
type Host struct {
	Name  string       `xml:"name,attr"`
	Addr  string       `xml:"addr,attr"`
	Port  int          `xml:"port,attr"`
	User  string       `xml:"user"`
	Key   string       `xml:"key,omitempty"`
	Pass  string       `xml:"password,omitempty"`
	Via   string       `xml:"via,omitempty"`
	SOCKS string       `xml:"socks,omitempty"`
	Other KeyValueList `xml:"other,omitempty"`
}

// Forward is a port forwarding directive
type Forward struct {
	Name    string        `xml:"name,attr"`
	Lines   []ForwardLine `xml:"line"`
	Comment string        `xml:",comment"`
}

// ForwardLine is a specific port or range or ports to forward
type ForwardLine struct {
	SrcIP   string `xml:"srcip,attr"`
	SrcPort int    `xml:"srcport,attr"`
	DstIP   string `xml:"dstip,attr"`
	DstPort int    `xml:"dstport,attr"`
	Repeat  int    `xml:"repeat,attr,omitempty"`
}

// KeyValue is a simple key = value binding
type KeyValue struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

// KeyValueList is a slice of KeyValues
type KeyValueList []KeyValue

// Get gets the value for a named key or the empty string
func (kvs KeyValueList) Get(key string) string {
	for _, kv := range kvs {
		if kv.Key == key {
			return kv.Value
		}
	}
	return ""
}

// SrcString returns the source IP address and port as a string formatted for
// use with Dial() and similar.
func (line ForwardLine) SrcString(i int) string {
	if i > line.Repeat {
		panic("index > repeat")
	}
	return fmt.Sprintf("%s:%d", line.SrcIP, line.SrcPort+i)
}

// DstString returns the destination IP address and port as a string formatted for
// use with Dial() and similar.
func (line ForwardLine) DstString(i int) string {
	if i > line.Repeat {
		panic("index > repeat")
	}
	return fmt.Sprintf("%s:%d", line.DstIP, line.DstPort+i)
}

// String returns a human readable representation of the port forward.
func (line ForwardLine) String() string {
	if line.Repeat == 0 {
		src := fmt.Sprintf("%s:%d", line.SrcIP, line.SrcPort)
		dst := fmt.Sprintf("%s:%d", line.DstIP, line.DstPort)
		return fmt.Sprintf("%s -> %s", src, dst)
	}
	src := fmt.Sprintf("%s:%d-%d", line.SrcIP, line.SrcPort, line.SrcPort+line.Repeat)
	dst := fmt.Sprintf("%s:%d-%d", line.DstIP, line.DstPort, line.DstPort+line.Repeat)
	return fmt.Sprintf("%s -> %s", src, dst)
}

// Load loads and parses an io.Reader as a tunnel config, returning a
// Config pointer or an error.
func Load(r io.Reader) (*Config, error) {
	return parse(ini.Parse(r))
}

// SourceAddresses returns all the source addresses used in forwarding
// directives.
func (c *Config) SourceAddresses() []string {
	addrMap := make(map[string]bool)
	for _, fwd := range c.Forwards {
		for _, line := range fwd.Lines {
			addrMap[line.SrcIP] = true
		}
	}

	addrs := make([]string, 0, len(addrMap))
	for ip := range addrMap {
		addrs = append(addrs, ip)
	}

	sort.Strings(addrs)
	return addrs
}

// Remap changes all forwarding directives to use the default localhost
// address 127.0.0.1 instead of their configured source, if it differs from
// 127.0.0.1.
func (c *Config) Remap() {
	port := 10000
	for fi := range c.Forwards {
		for li := range c.Forwards[fi].Lines {
			if c.Forwards[fi].Lines[li].SrcIP != "127.0.0.1" {
				// BUG: Need to keep track of used ports and not try to use them twice.
				c.Forwards[fi].Lines[li].SrcIP = "127.0.0.1"
				c.Forwards[fi].Lines[li].SrcPort = port
				port += c.Forwards[fi].Lines[li].Repeat + 1
			}
		}
	}
}

// FeatureFlags returns the set of features required to handle a tunnel
// configuration.
func (c *Config) FeatureFlags() uint32 {
	var flags uint32

	for _, h := range c.Hosts {
		if h.Key != "" {
			flags |= FeatureSshKey
		} else if h.Pass != "" {
			flags |= FeatureSshPassword
		}
	}

	if c.Vpnc != nil {
		flags |= FeatureVpnc
	}
	if c.OpenConnect != nil {
		flags |= FeatureOpenConnect
	}
	if c.General.Main != "" && c.GetHost(c.General.Main).SOCKS != "" {
		flags |= FeatureSocks
	}
	if len(c.Hosts) == 0 {
		flags |= FeatureLocalOnly
	}

	return flags
}

// GetHost returns a pointer to the named host or nil
func (c *Config) GetHost(name string) *Host {
	for _, host := range c.Hosts {
		if host.Name == name {
			return &host
		}
	}
	return nil
}

func (c *Config) WriteXML(w io.Writer) (int, error) {
	bs, err := xml.MarshalIndent(c, "", "  ")
	if err != nil {
		return 0, err
	}
	return w.Write(bs)
}
