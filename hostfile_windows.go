package main

import (
	"os"
)
var (
	HOSTFILE = os.Getenv("SystemRoot")+"/system32/drivers/etc/hosts"
	FLUSH_CMD []string = nil
)
