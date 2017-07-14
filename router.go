package main

import (
	"encoding/json"
	"net/http"
	"github.com/gorilla/mux"
)

type ApiError struct {
	Code int `json:"code"`
	Text string `json:"text"`
}

func Index(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, struct {
		App string `json:"app"`
		Name string `json:"name"`
		Version string `json:"version"`
	}{
		App: "enclave_server",
		Name: "Default Enclave Server",
		Version: serverversion,
	})
}

func WriteJSON(w http.ResponseWriter, val interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, _ := json.Marshal(val)
	w.Write(b)
}

func ApiRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		if route.Restricted {
			handler = ValidateRequest(handler)
		}

		router.Methods(route.Method).Path(route.Pattern).Name(route.Name).Handler(handler)
	}

	return router
}