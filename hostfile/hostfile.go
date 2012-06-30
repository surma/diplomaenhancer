package hostfile

import (
	"regexp"
	"strings"
)

type Hostfile []Block

type Block struct {
	Comment []string
	Entries []Entry
}

type Entry struct {
	IP        string
	Hostnames []string
}

func (h Hostfile) String() string {
	sep := ""
	buf := ""
	for _, block := range h {
		buf += sep
		sep = "\n"
		if len(block.Comment) > 0 {
			buf += "#" + strings.Join(block.Comment, "\n#") + "\n"
		}
		for _, entry := range block.Entries {
			buf += entry.IP + " " + strings.Join(entry.Hostnames, " ") + "\n"
		}
	}
	return buf
}

var (
	ipMatcher = regexp.MustCompile("^([0-9]{1,3}\\.){3}[0-9]{1,3}$")
)

func (t Entry) Valid() bool {
	return ipMatcher.MatchString(t.IP) && len(t.Hostnames) >= 1 && len(t.Hostnames[0]) >= 1
}
