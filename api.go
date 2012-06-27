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

func conditionalFailf(w http.ResponseWriter, cond bool, code int, sfmt string, v ...interface{}) bool {
	if cond {
		err := fmt.Sprintf(sfmt, v...)
		http.Error(w, err, code)
		log.Println(err)
	}
	return cond
}

func serveAPI(addr string) {
	r := mux.NewRouter()
	apirouter := r.PathPrefix("/api").Subrouter()
	r.PathPrefix("/admin").Handler(http.StripPrefix("/admin", http.FileServer(http.Dir("./admin"))))
	apirouter.Path("/").Methods("GET").HandlerFunc(apiListHandler)
	apirouter.Path("/state").Methods("POST").HandlerFunc(authenticationWrapper(apiStateHandler))
	apirouter.Path("/{ip:[0-9.]+}").Methods("GET").HandlerFunc(apiListHostHandler)
	apirouter.Path("/{ip:[0-9.]+}").Methods("POST").HandlerFunc(apiAddHostHandler)
	apirouter.Path("/{ip:[0-9.]+}").Methods("DELETE").HandlerFunc(authenticationWrapper(apiDeleteHostHandler))
	e := http.ListenAndServe(addr, r)
	if e != nil {
		log.Fatalf("serveAPI: Could not bind http server: %s", e)
	}
}

func apiListHandler(w http.ResponseWriter, r *http.Request) {
	d, e := json.Marshal(hosts)
	if conditionalFailf(w, e != nil, http.StatusInternalServerError, "List: Could not marshal hosts: %s", e) { return }

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(d)))
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func apiListHostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if conditionalFailf(w, hosts.HasIP(vars["ip"]), http.StatusNotFound, "ListHost: Undefined ip") { return }

	d, e := json.Marshal(hosts[vars["ip"]])
	if conditionalFailf(w, e != nil, http.StatusInternalServerError, "ListHost: Could not marshal hosts: %s", e) { return }

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(d)))
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func apiDeleteHostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	line, _, e := bufio.NewReader(r.Body).ReadLine()
	if conditionalFailf(w, e != nil, http.StatusBadRequest, "DeleteHost: Received invalid request: %s", e) { return }

	hostnames := strings.Fields(string(line))
	if conditionalFailf(w, len(hostnames) > 1, http.StatusBadRequest, "DeleteHost: Received multiple fields") { return }

	e = hosts.Remove(vars["ip"], hostnames[0])
	if conditionalFailf(w, e != nil, http.StatusBadRequest, "DeleteHost: Could not remove host: %s", e) { return }

	if active {
		hosts.WriteToFile(HOSTSFILE)
	}
	w.WriteHeader(http.StatusNoContent)
}

func apiAddHostHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	line, _, e := bufio.NewReader(r.Body).ReadLine()
	if conditionalFailf(w, e != nil, http.StatusBadRequest, "AddHost: Received invalid request: %s", e) { return }

	hostnames := strings.Fields(string(line))
	hosts.AddMultiple(vars["ip"], hostnames)
	if active {
		hosts.WriteToFile(HOSTSFILE)
	}
	w.WriteHeader(http.StatusNoContent)
}

func apiStateHandler(w http.ResponseWriter, r *http.Request) {
	line, _, e := bufio.NewReader(r.Body).ReadLine()
	if conditionalFailf(w, e != nil, http.StatusBadRequest, "State: Received invalid request: %s", e) { return }

	state := strings.TrimSpace(string(line))
	switch state {
	case "active":
		hosts.WriteToFile(HOSTSFILE)
		active = true
	case "inactive":
		restoreHostsFile(backup)
		active = false
	default:
		conditionalFailf(w, true, http.StatusBadRequest, "State: Received invalid state: %s", state)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func authenticationWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		if r.Header.Get("X-DiplomaEnhancer-Token") != password {
			log.Printf("Received invalid password")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h(w, r)
	}

}

