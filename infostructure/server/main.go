package server

import (
	"fmt"
	"net/http"
	"strings"
)

func Run(host string, port int, logURL string) error {
	addr := fmt.Sprintf("%s:%d", host, port)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("octopus"))

		go sendLog(logURL, fmt.Sprintf("GET / from %s", r.RemoteAddr))
	})

	return http.ListenAndServe(addr, nil)
}

func sendLog(url, message string) {
	_, _ = http.Post(url, "text/plain", strings.NewReader(message))
}
