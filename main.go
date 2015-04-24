package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"unicode"
)

var client = new(http.Client)

func httpHandler(w http.ResponseWriter, r *http.Request) {
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Printf("error dumping request: %s\n", err)
		return
	}
	os.Stdout.Write(b)

	r.RequestURI = ""
	r.URL.Scheme = strings.Map(unicode.ToLower, r.URL.Scheme)

	res, err := client.Do(r)
	if err != nil {
		fmt.Printf("error forwarding request: %s\n", err)
		return
	}
	defer res.Body.Close()

	for k, v := range res.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(res.StatusCode)

	if _, err := io.Copy(w, res.Body); err != nil {
		fmt.Printf("error copying body: %s\n", err)
	}
}

func main() {
	http.HandleFunc("/", httpHandler)
	http.ListenAndServe(":8080", nil)
}
