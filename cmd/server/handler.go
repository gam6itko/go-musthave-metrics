package main

import "net/http"

// getPingHandler метод для проверки работы сервера.
func getPingHandler(resp http.ResponseWriter, req *http.Request) {
	err := Database.PingContext(req.Context())
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-Type", "text/html")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("OK"))
}
