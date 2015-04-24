package main

import (
	"flag"
	"fmt"
	"github.com/jordanorelli/moon/lib"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"unicode"
)

var client = new(http.Client)

func httpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("from:", r.RemoteAddr)
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
	if _, ok := w.Header()["Proxy-Connection"]; ok {
		delete(w.Header(), "Proxy-Connection")
	}
	w.WriteHeader(res.StatusCode)

	if _, err := io.Copy(w, res.Body); err != nil {
		fmt.Printf("error copying body: %s\n", err)
	}
}

func bail(status int, t string, args ...interface{}) {
	if status == 0 {
		fmt.Fprintf(os.Stdout, t+"\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, t+"\n", args...)
	}
	os.Exit(status)
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./prox_config.moon", "path to configuration file")
	flag.Parse()

	conf, err := moon.ReadFile(configPath)
	if err != nil {
		bail(1, "unable to read config: %s", err)
	}

	var addr string
	if err := conf.Get("proxy_addr", &addr); err != nil {
		bail(1, "error reading proxy_addr from config: %s", err)
	}

	http.HandleFunc("/", httpHandler)
	http.ListenAndServe(addr, nil)
}
