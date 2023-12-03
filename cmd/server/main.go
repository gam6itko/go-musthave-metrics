package main

import (
	updateAction "github.com/gam6itko/go-musthave-metrics/internal/controller/update"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", updateAction.Handler)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
