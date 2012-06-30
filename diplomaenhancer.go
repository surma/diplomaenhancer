package main

import (
	"./hostfile"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	PASSWORD = "apassword"
)

var (
	api  = flag.String("api", "0.0.0.0:13370", "Address to bind the API interface to")
	help = flag.Bool("help", false, "Show this help")
)

var (
	originalhostfile hostfile.Hostfile
	password         string
	active           bool = false
)

func main() {
	flag.Parse()

	if *help {
		fmt.Println("Usage: diplomaenhancer [options]")
		flag.PrintDefaults()
		return
	}

	e := readHostfile()
	if e != nil {
		log.Fatalf("Could not manipulate hosts file %s: %s", HOSTFILE, e)
	}

	e = readBlocklist()
	if e != nil {
		log.Printf("Could not blocklist file: %s. Creating...", e)
	}

	log.Printf("Starting server...")
	password = PASSWORD
	go serveBlockpage()
	serveAPI(*api)
}

func readHostfile() error {
	// Check for write permissions
	f, e := os.OpenFile(HOSTFILE, os.O_RDWR, os.FileMode(0644))
	if e != nil {
		return e
	}
	defer f.Close()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		<-c
		restoreHostfile()
		os.Exit(0)
	}()

	originalhostfile, e = hostfile.Parse(f)
	return e
}

func restoreHostfile() {
	f, e := os.Create(HOSTFILE)
	if e != nil {
		log.Fatalf("Could not restore host file %s: %s", HOSTFILE, e)
	}
	defer f.Close()
	f.Write([]byte(originalhostfile.String()))
}
