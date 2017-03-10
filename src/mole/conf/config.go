package conf

import (
	"fmt"
	"io"
	"net"
	"sort"

	"github.com/calmh/ini"
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
	Comments []string

	General struct {
		Description string
		Author      string
		Main        string
		Version     int
		Other       map[string]string
		Comments    []string
	}

	Hosts       []Host
	Forwards    []Forward
	HostsMap    map[string]int
	OpenConnect map[string]string
	Vpnc        map[string]string
	VpnRoutes   []string
}

// Host is an SSH host to bounce via
type Host struct {
	Name     string
	Addr     string
	Port     int
	User     string
	Key      string
	Pass     string
	Via      string
	SOCKS    string
	Other    map[string]string
	Comments []string
}

// Forward is a port forwarding directive
type Forward struct {
	Name     string
	Lines    []ForwardLine
	Other    map[string]string
	Comments []string
}

// Addr is a composite type of IPAddr and TCP ports
type Addrports struct {
	Addr  net.IP
	Ports []int
}

// ForwardLine is a specific port or range or ports to forward
type ForwardLine struct {
	Src Addrports
	Dst Addrports
	//SrcIP   string
	//SrcPort int
	//DstIP   string
	//DstPort int
	//Repeat  int
}

// SrcString returns the source IP address and port as a string formatted for
// use with Dial() and similar.
func (line ForwardLine) SrcString(i int) string {
	if i >= len(line.Src.Ports) {
		panic("index > repeat")
	}
	if line.Src.Addr.To4() != nil {
		return fmt.Sprintf("%s:%d", line.Src.Addr.String(), line.Src.Ports[i])
	} else {
		return fmt.Sprintf("[%s]:%d", line.Src.Addr.String(), line.Src.Ports[i])
	}
}

// DstString returns the destination IP address and port as a string formatted for
// use with Dial() and similar.
func (line ForwardLine) DstString(i int) string {
	//if i > line.Repeat {
	if i >= len(line.Dst.Ports) {
		panic("index > repeat")
	}
	if line.Dst.Addr.To4() != nil {
		return fmt.Sprintf("%s:%d", line.Dst.Addr.String(), line.Dst.Ports[i])
	} else {
		return fmt.Sprintf("[%s]:%d", line.Dst.Addr.String(), line.Dst.Ports[i])
	}
}

// String returns a human readable representation of the port forward.
func (line ForwardLine) String() string {
	if len(line.Src.Ports) == 1 {
		src := fmt.Sprintf("%s:%d", line.Src.Addr.String(), line.Src.Ports[0])
		dst := fmt.Sprintf("%s:%d", line.Dst.Addr.String(), line.Dst.Ports[0])
		return fmt.Sprintf("%s -> %s", src, dst)
	}
	src := fmt.Sprintf("%s:%d-%d", line.Src.Addr.String(), line.Src.Ports[0], line.Src.Ports[len(line.Src.Ports)-1])
	dst := fmt.Sprintf("%s:%d-%d", line.Dst.Addr.String(), line.Dst.Ports[0], line.Dst.Ports[len(line.Src.Ports)-1])
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
			addrMap[line.Src.Addr.String()] = true
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
			if c.Forwards[fi].Lines[li].Src.Addr.String() != "127.0.0.1" && c.Forwards[fi].Lines[li].Src.Addr.String() != "[::1]" {
				// BUG: Need to keep track of used ports and not try to use them twice.
				c.Forwards[fi].Lines[li].Src.Addr = net.ParseIP("127.0.0.1")
				for sp := range c.Forwards[fi].Lines[li].Src.Ports {
					c.Forwards[fi].Lines[li].Src.Ports[sp] = port
					port += 1
				}
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
	if c.General.Main != "" && c.Hosts[c.HostsMap[c.General.Main]].SOCKS != "" {
		flags |= FeatureSocks
	}
	if len(c.Hosts) == 0 {
		flags |= FeatureLocalOnly
	}

	return flags
}
