package main

import (
	"./hostfile"
	"bufio"
	"code.google.com/p/gorilla/mux"
	"encoding/json"
	"fmt"
	"github.com/surma/gouuid"
	"github.com/surma/goappdata"
	"log"
	"net/http"
	"os"
	"os/exec"
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
	r.Path("/").Handler(http.RedirectHandler("/admin/", http.StatusMovedPermanently))
	apirouter := r.PathPrefix("/api").Subrouter()
	apirouter.Path("/").Methods("GET").HandlerFunc(apiListHandler)
	apirouter.Path("/").Methods("POST").HandlerFunc(updateWrapper(apiAddHandler))
	apirouter.Path("/state").Methods("POST").HandlerFunc(authenticationWrapper(updateWrapper(apiSetStateHandler)))
	apirouter.Path("/state").Methods("GET").HandlerFunc(apiGetStateHandler)
	apirouter.Path("/{uuid:[0-9a-f-]+}").Methods("GET").HandlerFunc(apiListSingleHandler)
	apirouter.Path("/{uuid:[0-9a-f-]+}").Methods("DELETE").HandlerFunc(authenticationWrapper(updateWrapper(apiDeleteSingleHandler)))

	r.PathPrefix("/admin/").Handler(http.StripPrefix("/admin", http.FileServer(http.Dir("./admin"))))

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

func apiSetStateHandler(w http.ResponseWriter, r *http.Request) {
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

func apiGetStateHandler(w http.ResponseWriter, r *http.Request) {
	if active {
		w.Write([]byte("active"))
	} else {
		w.Write([]byte("inactive"))
	}
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

func updateWrapper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r)
		newhostfile := make([]hostfile.Block, len(originalhostfile))
		copy(newhostfile, originalhostfile)
		if active {
			block := hostfile.Block {
				Comment: []string{"DiplomaEnhancer:"},
				Entries: make([]hostfile.Entry, 0, len(blocklist)),
			}
			for _, entry := range blocklist {
				block.Entries = append(block.Entries, entry)
			}
			newhostfile = append(newhostfile, block)
		}

		e := saveBlocklist()
		if e != nil {
			log.Printf("Could not save to blocklist file: %s", e)
		}
		e = writeHostfile(hostfile.Hostfile(newhostfile))
		if e != nil {
			log.Fatalf("Could not write host file: %s")
		}
		e = exec.Command(FLUSH_CMD[0], FLUSH_CMD[1:]...).Start()
		if e != nil {
			log.Printf("Could not flush dns cache: %s", e)
		}
	}
}

func saveBlocklist() error {
	blocklistfile, e := goappdata.CreatePath("diplomaenhancer")
	if e != nil {
		return e
	}
	blocklistfile += "/blocklist"
	f, e := os.Create(blocklistfile)
	if e != nil {
		return e
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	e = enc.Encode(blocklist)
	return e
}

func readBlocklist() error {
	blocklistfile, e := goappdata.CreatePath("diplomaenhancer")
	if e != nil {
		return e
	}
	blocklistfile += "/blocklist"
	f, e := os.Open(blocklistfile)
	if e != nil {
		return e
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	e = dec.Decode(&blocklist)
	return e
}

func writeHostfile(hostfile hostfile.Hostfile) error {
	f, e := os.Create(HOSTFILE)
	if e != nil {
		return e
	}
	defer f.Close()
	_, e = f.Write([]byte(hostfile.String()))
	return e
}
