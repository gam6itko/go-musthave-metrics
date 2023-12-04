package main

import (
	updateAction "github.com/gam6itko/go-musthave-metrics/internal/server/controller/update"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", updateAction.Handle)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
