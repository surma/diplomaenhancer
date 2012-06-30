package main

const (
	HOSTFILE = "/etc/hosts"
)

var (
	FLUSH_CMD = []string{"dscacheutil", "-flushcache"}
)
