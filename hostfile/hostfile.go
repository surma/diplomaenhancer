package hostfile

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Hostfile []Block

type Block struct {
	Comment string
	IP string
	Hostnames []string
}

func (h Hosts) Add(ip, hostname string) {
	if _, ok := h[ip]; !ok {
		h[ip] = make([]string, 0, 2)
	}
	h[ip] = append(h[ip], hostname)
}

func (h Hosts) HasIP(ip string) bool {
	_, ok := h[ip]
	return ok
}

func (h Hosts) Remove(ip, hostname string) error {
	if _, ok := h[ip]; !ok {
		return fmt.Errorf("Unknown ip")
	}
	newhostnames := make([]string, 0, 2)
	for _, ohostname := range h[ip] {
		if hostname == ohostname {
			continue
		}
		newhostnames = append(newhostnames, ohostname)
	}
	if len(newhostnames) <= 0 {
		delete(h, ip)
	} else {
		h[ip] = newhostnames
	}
	return nil
}

func (h Hosts) AddMultiple(ip string, hostnames []string) {
	if _, ok := h[ip]; !ok {
		h[ip] = make([]string, 0, 2)
	}
	h[ip] = append(h[ip], hostnames...)
}

func (h Hosts) WriteToFile(path string) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	h.Write(f)
	return nil
}

func (h Hosts) Write(w io.Writer) {
	for ip, hostnames := range h {
		fmt.Fprintf(w, "%s ", ip)
		for _, hostname := range hostnames {
			fmt.Fprintf(w, "%s ", hostname)
		}
		fmt.Fprintf(w, "\n")
	}
}
