package main

import (
	"fmt"
	"net/http"
)

func requestsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select * from requests limit 100")
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to query db: %s", err), 500)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id   string
			host string
			path string
		)
		rows.Scan(&id, &host, &path)
		fmt.Fprintf(w, "%s %s %s\n", id, host, path)
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
