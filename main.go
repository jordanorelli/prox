package main

import (
	"flag"
	"fmt"
	"github.com/jordanorelli/moon/lib"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"
	"unicode"
)

var (
	client = new(http.Client)
	conf   *moon.Doc
)

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
	var addr string
	if err := conf.Get("proxy_addr", &addr); err != nil {
		bail(1, "error reading proxy_addr from config: %s", err)
	}

	m := http.NewServeMux()
	m.HandleFunc("/", httpHandler)
	http.ListenAndServe(addr, m)
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./prox_config.moon", "path to configuration file")
	flag.Parse()

	log.Printf("reading config from %s", configPath)
	var err error
	conf, err = moon.ReadFile(configPath)
	if err != nil {
		bail(1, "unable to read config: %s", err)
	}

	if err := openDB(); err != nil {
		bail(1, "unable to open db: %s", err)
	}
	defer db.Close()

	go appServer()
	proxyListener()
}
