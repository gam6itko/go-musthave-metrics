package main

import (
	"log"
	"net/http"
)

// getPingHandler метод для проверки работы сервера.
func getPingHandler(resp http.ResponseWriter, req *http.Request) {
	err := Database.PingContext(req.Context())
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-Type", "text/html")
	resp.WriteHeader(http.StatusOK)
	if _, err2 := resp.Write([]byte("OK")); err2 != nil {
		log.Fatal("Failed to write response: ", err2)
	}
}
