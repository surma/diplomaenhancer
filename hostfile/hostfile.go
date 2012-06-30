package hostfile

import (
	"strings"
)

type Hostfile []Block

type Block struct {
	Comment []string
	Entries []Entry
}

type Entry struct {
	IP string
	Hostnames []string
}

func (h Hostfile) String() string {
	sep := ""
	buf := ""
	for _, block := range h {
		buf += sep
		sep = "\n"
		if len(block.Comment) > 0 {
			buf += "#"+strings.Join(block.Comment, "\n#")+"\n"
		}
		for _, entry := range block.Entries {
			buf += entry.IP + " " + strings.Join(entry.Hostnames, " ")+"\n"
		}
	}
	return buf
}

