// +build ignore

package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func newRouterInstances() {
	r := mux.NewRouter()
	r.Handle("/new/gorillamux/handle", nil)
	r.HandleFunc("/new/gorillamux/handlefunc", nil)

	m := http.NewServeMux()
	m.Handle("/new/httpmux/handle", nil)
	m.HandleFunc("/new/httpmux/handlefunc", nil)
}

func routerAsArguments(r *mux.Router, m *http.ServeMux) {
	r.Handle("/arg/gorillamux/handle", nil)
	r.HandleFunc("/arg/gorillamux/handlefunc", nil)
	m.Handle("/arg/httpmux/handle", nil)
	m.HandleFunc("/arg/httpmux/handlefunc", nil)
}
