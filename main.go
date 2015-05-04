package main

import (
	"fmt"
	"github.com/jordanorelli/moon/lib"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
	"unicode"
)

var (
	client = new(http.Client)
)

var config struct {
	ProxyAddr string `
    name: proxy_addr
    default: ":8080"
    help: proxy address. Browsers send their http traffic to this port.
    `

	AppAddr string `
    name: app_addr
    default: ":9000"
    help: app address. Users visit this address to view the proxy's history db
    `

	DbPath string `
    name: dbpath
    default: history.db
    help: path to a sqlite file used for storing the user's history
    `
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	id := newRequestId()
	fmt.Printf("%s >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n", id.String())
	fmt.Println("from:", r.RemoteAddr)
	if err := freezeRequest(r); err != nil {
		fmt.Printf("error freezing request: %s\n", err)
		return
	}
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Printf("error dumping request: %s\n", err)
		return
	}
	os.Stdout.Write(b)

	requestURI := r.RequestURI
	r.RequestURI = ""
	r.URL.Scheme = strings.Map(unicode.ToLower, r.URL.Scheme)

	start := time.Now()
	res, err := client.Do(r)
	if err != nil {
		fmt.Printf("error forwarding request: %s\n", err)
		return
	}
	defer res.Body.Close()

	fmt.Printf("%s <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n", id.String())
	resb, err := httputil.DumpResponse(res, false)
	if err != nil {
		fmt.Printf("fuuuuuuuuuck")
	}
	os.Stdout.Write(resb)

	for k, v := range res.Header {
		w.Header()[k] = v
	}
	if _, ok := w.Header()["Proxy-Connection"]; ok {
		delete(w.Header(), "Proxy-Connection")
	}
	w.WriteHeader(res.StatusCode)

	if _, err := io.Copy(w, res.Body); err != nil {
		fmt.Printf("error copying body: %s\n", err)
	}
	elapsed := time.Since(start)

	r.RequestURI = requestURI
	fmt.Printf("elapsed: %v (%v)\n", elapsed, elapsed.Nanoseconds()/1000000)
	saveRequest(id, r)
	saveResponse(id, res, elapsed)
}

func bail(status int, t string, args ...interface{}) {
	if status == 0 {
		fmt.Fprintf(os.Stdout, t+"\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, t+"\n", args...)
	}
	os.Exit(status)
}

func proxyListener() {
	m := http.NewServeMux()
	m.HandleFunc("/", httpHandler)
	http.ListenAndServe(config.ProxyAddr, m)
}

func main() {
	moon.Parse(&config)

	if err := openDB(); err != nil {
		bail(1, "unable to open db: %s", err)
	}
	defer db.Close()

	go appServer()
	proxyListener()
}
