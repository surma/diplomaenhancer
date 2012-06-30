package main

import (
	"./hostfile"
	"bufio"
	"code.google.com/p/gorilla/mux"
	"encoding/json"
	"fmt"
	"github.com/surma/gouuid"
	"log"
	"net/http"
	"strings"
)

var (
	blocklist = map[string]hostfile.Entry{}
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
	apirouter.Path("/").Methods("POST").HandlerFunc(apiAddHandler)
	apirouter.Path("/state").Methods("POST").HandlerFunc(authenticationWrapper(apiStateHandler))
	apirouter.Path("/{uuid:[0-9a-f-]+}").Methods("GET").HandlerFunc(apiListSingleHandler)
	apirouter.Path("/{uuid:[0-9a-f-]+}").Methods("DELETE").HandlerFunc(authenticationWrapper(apiDeleteSingleHandler))
	e := http.ListenAndServe(addr, r)
	if e != nil {
		log.Fatalf("serveAPI: Could not bind http server: %s", e)
	}
}

func apiListHandler(w http.ResponseWriter, r *http.Request) {
	d, e := json.Marshal(blocklist)
	if conditionalFailf(w, e != nil, http.StatusInternalServerError, "List: Could not marshal hosts: %s", e) {
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(d)))
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func apiAddHandler(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	entry := hostfile.Entry{}
	e := d.Decode(&entry)
	if conditionalFailf(w, e != nil, http.StatusBadRequest, "Add: Received invalid request: %s", e) {
		return
	}
	if conditionalFailf(w, !entry.Valid(), http.StatusBadRequest, "Add: Received invalid entry") {
		return
	}

	uuid := gouuid.New()
	blocklist[uuid.String()] = entry
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(uuid.String()))
}

func apiListSingleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entry, ok := blocklist[vars["uuid"]]
	if conditionalFailf(w, !ok, http.StatusNotFound, "ListSingle: Unknown UUID %s", vars["uuid"]) {
		return
	}

	d, e := json.Marshal(entry)
	if conditionalFailf(w, e != nil, http.StatusInternalServerError, "ListSingle: Could not marshal hosts: %s", e) {
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(d)))
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func apiDeleteSingleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_, ok := blocklist[vars["uuid"]]
	if conditionalFailf(w, !ok, http.StatusNotFound, "DeleteSingle: Unknown UUID %s", vars["uuid"]) {
		return
	}

	delete(blocklist, vars["uuid"])
	w.WriteHeader(http.StatusNoContent)
}

func apiStateHandler(w http.ResponseWriter, r *http.Request) {
	line, _, e := bufio.NewReader(r.Body).ReadLine()
	if conditionalFailf(w, e != nil, http.StatusBadRequest, "State: Received invalid request: %s", e) {
		return
	}

	state := strings.TrimSpace(string(line))
	switch state {
	case "active":
		active = true
	case "inactive":
		active = false
	default:
		conditionalFailf(w, true, http.StatusBadRequest, "State: Received invalid state: %s", state)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func authenticationWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-DiplomaEnhancer-Token") != password {
			log.Printf("Received invalid password")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h(w, r)
	}
}
