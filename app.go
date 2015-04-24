package main

import (
	"fmt"
	"net/http"
)

func requestsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("we have %d requests in history when user checked", len(requestHistory))
	for _, req := range requestHistory {
		fmt.Fprintln(w, req.RequestURI)
	}
}

func appServer() {
	var addr string
	if err := conf.Get("app_addr", &addr); err != nil {
		bail(1, "error reading app_addr from config: %s", err)
	}

	m := http.NewServeMux()
	m.HandleFunc("/requests", requestsHandler)
	http.ListenAndServe(addr, m)
}
