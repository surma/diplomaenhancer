package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Hosts map[string][]string

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

func New() Hosts {
	return map[string][]string{}
}

func ParseString(content string) (Hosts, error) {
	return Parse(strings.NewReader(content))
}

func Parse(r io.Reader) (Hosts, error) {
	h := New()
	b := bufio.NewReader(r)
	for bline, prefix, e := b.ReadLine(); e == nil; bline, prefix, e = b.ReadLine() {
		for prefix {
			var blinerest []byte
			blinerest, prefix, e = b.ReadLine()
			bline = append(bline, blinerest...)
			if e != nil {
				break
			}
		}

		line := strings.TrimSpace(string(bline))
		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 0 {
			return nil, fmt.Errorf("Invalid line: \"%s\"", line)
		}
		h.AddMultiple(fields[0], fields[1:])
	}
	return h, nil
}
