package main

import (
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	err := http.ListenAndServe(`:8080`, newRouter())
	if err != nil {
		panic(err)
	}
}

func newRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", getAllMetrics)
	r.Get("/value/{type}/{name}", getValueHandler)
	r.Post("/update/{type}/{name}/{value}", postUpdateHandler)

	return r
}

func getAllMetrics(resp http.ResponseWriter, req *http.Request) {
	for name, val := range memory.CounterAll() {
		io.WriteString(resp, fmt.Sprintf("%s: %d\n", name, val))
	}
	for name, val := range memory.GaugeAll() {
		io.WriteString(resp, fmt.Sprintf("%s: %f\n", name, val))
	}
}

func getValueHandler(resp http.ResponseWriter, req *http.Request) {
	// fmt.Printf("requst: [%s] %s\n", req.Method, req.URL)

	name := chi.URLParam(req, "name")
	if name == "" {
		http.Error(resp, "Bad name", http.StatusNotFound)
		return
	}

	switch chi.URLParam(req, "type") {
	case "counter":
		val, exists := memory.CounterGet(name)
		if !exists {
			http.Error(resp, "Not found", http.StatusNotFound)
			return
		}
		io.WriteString(resp, fmt.Sprintf("%d", val))

	case "gauge":
		val, exists := memory.GaugeGet(name)
		if !exists {
			http.Error(resp, "Not found", http.StatusNotFound)
			return
		}
		io.WriteString(resp, fmt.Sprintf("%g", val))

	default:
		http.Error(resp, "invalid metric type", http.StatusNotFound)
		return
	}
}

func postUpdateHandler(resp http.ResponseWriter, req *http.Request) {
	//fmt.Printf("requst: [%s] %s\n", req.Method, req.URL)

	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	switch strings.ToLower(chi.URLParam(req, "type")) {
	case "counter":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(resp, "invalid counter value", http.StatusBadRequest)
			return
		}
		memory.CounterInc(name, v)

	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(resp, "invalid gauge value", http.StatusBadRequest)
			return
		}
		memory.GaugeSet(name, v)

	default:
		http.Error(resp, "invalid metric type", http.StatusBadRequest)
		return
	}

	resp.WriteHeader(http.StatusOK)
	io.WriteString(resp, "OK")
}
