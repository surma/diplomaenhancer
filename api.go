package main

import (
	"bufio"
	"code.google.com/p/gorilla/mux"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var (
	hosts Hosts
	password string
)

func serveAPI(addr string, _hosts Hosts, _password string) {
	// I'm so sorry for this
	hosts = _hosts
	password = _password

	r := mux.NewRouter()
	apirouter := r.PathPrefix("/api").Subrouter()
	adminrouter := r.PathPrefix("/admin").Subrouter()
	apirouter.Path("/").Methods("GET").HandlerFunc(apiListHandler)
	apirouter.Path("/{ip:[0-9.:%]+}").Methods("GET").HandlerFunc(apiListHostHandler)
	apirouter.Path("/{ip:[0-9.:%]+}").Methods("POST").HandlerFunc(apiAddHostHandler)
	adminrouter.Methods("GET").HandlerFunc(adminhandler)
	e := http.ListenAndServe(addr, r)
	if e != nil {
		log.Fatalf("Could not bind http server: %s", e)
	}
}

func apiListHandler(w http.ResponseWriter, r *http.Request) {
	d, e := json.Marshal(hosts)
	if e != nil {
		log.Fatalf("Could not marshal hosts: %s", e)
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(d)))
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func apiListHostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if _, ok := hosts[vars["ip"]]; !ok {
		http.Error(w, "Undefined ip", http.StatusNotFound)
		return
	}

	d, e := json.Marshal(hosts[vars["ip"]])
	if e != nil {
		log.Fatalf("Could not marshal hosts: %s", e)
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(d)))
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func apiAddHostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-DiplomaEnhancer-Token") != password {
		w.WriteHeader(401)
		return
	}
	vars := mux.Vars(r)
	line, _, e := bufio.NewReader(r.Body).ReadLine()
	if e != nil {
		log.Printf("Received invalid request: %s", e)
		w.WriteHeader(400)
		return
	}
	hostnames := strings.Fields(string(line))
	hosts.AddMultiple(vars["ip"], hostnames)
	hosts.WriteToFile(HOSTSFILE)
	w.WriteHeader(http.StatusNoContent)
}

func adminhandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ADMIN")
}
