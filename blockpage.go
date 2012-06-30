package main

import (
	"code.google.com/p/gorilla/mux"
	"fmt"
	"net/http"
	"log"
)

func serveBlockpage() {
	r := mux.NewRouter()
	r.PathPrefix("/iframe/").Handler(http.StripPrefix("/iframe", http.FileServer(http.Dir("./blockpage"))))
	r.PathPrefix("/").HandlerFunc(iframeHandler)
	e := http.ListenAndServe(":80", r)
	if e != nil {
		log.Fatalf("serveBlockpage: Could not bind http server: %s", e)
	}
}

func iframeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<body style="margin: 0; border: 0; padding: 0;"><iframe style="margin: 0; border: 0; padding: 0; width: 100%%; height: 100%%;" src="/iframe/index.html"></iframe>`)
}
