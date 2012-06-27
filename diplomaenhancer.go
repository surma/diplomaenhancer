package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	api  = flag.String("api", "0.0.0.0:13370", "Address to bind the API interface to")
	help = flag.Bool("help", false, "Show this help")
)

func main() {
	flag.Parse()

	if *help {
		fmt.Println("Usage: diplomaenhancer [options]")
		flag.PrintDefaults()
		return
	}

	backup, e := backupHostsFile()
	if e != nil {
		log.Fatalf("Could not manipulate hosts file %s: %s", HOSTSFILE, e)
	}
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		<-c
		restoreHostsFile(backup)
		os.Exit(0)
	}()

	hosts, e := ParseString(backup)
	if e != nil {
		log.Fatalf("Could not parse hosts file %s: %s", HOSTSFILE, e)
	}
	hosts.WriteToFile(HOSTSFILE)

	log.Printf("Starting server...")
	serveAPI(*api, hosts)
}

func backupHostsFile() (string, error) {
	// Check for write permissions
	f, e := os.OpenFile(HOSTSFILE, os.O_WRONLY, os.FileMode(0644))
	if e != nil {
		return "", e
	}
	f.Close()

	data, e := ioutil.ReadFile(HOSTSFILE)
	return string(data), e
}

func restoreHostsFile(content string) {
	f, e := os.Create(HOSTSFILE)
	if e != nil {
		log.Fatalf("Could not restore host file %s: %s", HOSTSFILE, e)
	}
	defer f.Close()
	f.Write([]byte(content))
}
