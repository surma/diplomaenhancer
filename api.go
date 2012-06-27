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

func serveAPI(addr string) {
	r := mux.NewRouter()
	apirouter := r.PathPrefix("/api").Subrouter()
	r.PathPrefix("/admin").Handler(http.StripPrefix("/admin", http.FileServer(http.Dir("./admin"))))
	apirouter.Path("/").Methods("GET").HandlerFunc(apiListHandler)
	apirouter.Path("/state").Methods("POST").HandlerFunc(apiStateHandler)
	apirouter.Path("/{ip:[0-9.]+}").Methods("GET").HandlerFunc(apiListHostHandler)
	apirouter.Path("/{ip:[0-9.]+}").Methods("POST").HandlerFunc(apiAddHostHandler)
	apirouter.Path("/{ip:[0-9.]+}").Methods("DELETE").HandlerFunc(apiDeleteHostHandler)
	e := http.ListenAndServe(addr, r)
	if e != nil {
		log.Fatalf("serveAPI: Could not bind http server: %s", e)
	}
}

func apiListHandler(w http.ResponseWriter, r *http.Request) {
	d, e := json.Marshal(hosts)
	if e != nil {
		log.Fatalf("List: Could not marshal hosts: %s", e)
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
		log.Fatalf("ListHost: Could not marshal hosts: %s", e)
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(d)))
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func apiDeleteHostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-DiplomaEnhancer-Token") != password {
		log.Printf("Received invalid password")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	line, _, e := bufio.NewReader(r.Body).ReadLine()
	if e != nil {
		log.Printf("DeleteHost: Received invalid request: %s", e)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hostnames := strings.Fields(string(line))
	if len(hostnames) > 1 {
		log.Printf("DeleteHost: Received multiple fields")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	e = hosts.Remove(vars["ip"], hostnames[0])
	if e != nil {
		log.Printf("DeleteHost: Could not remove host: %s", e)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if active {
		hosts.WriteToFile(HOSTSFILE)
	}
	w.WriteHeader(http.StatusNoContent)
}

func apiAddHostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-DiplomaEnhancer-Token") != password {
		log.Printf("AddHost: Received invalid password")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	line, _, e := bufio.NewReader(r.Body).ReadLine()
	if e != nil {
		log.Printf("AddHost: Received invalid request: %s", e)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hostnames := strings.Fields(string(line))
	hosts.AddMultiple(vars["ip"], hostnames)
	if active {
		hosts.WriteToFile(HOSTSFILE)
	}
	w.WriteHeader(http.StatusNoContent)
}

func apiStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-DiplomaEnhancer-Token") != password {
		log.Printf("State: Received invalid password")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	line, _, e := bufio.NewReader(r.Body).ReadLine()
	if e != nil {
		log.Printf("State: Received invalid request: %s", e)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	state := strings.TrimSpace(string(line))
	switch state {
	case "active":
		hosts.WriteToFile(HOSTSFILE)
		active = true
	case "inactive":
		restoreHostsFile(backup)
		active = false
	default:
		log.Printf("State: Received invalid state: %s", state)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
