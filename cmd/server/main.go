package main

import (
	updateAction "github.com/gam6itko/go-musthave-metrics/internal/controller/update"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", updateAction.Handler)
	http.ListenAndServe(`:8080`, mux)
}
