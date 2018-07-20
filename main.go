package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// cors adds required headers to responses such that direct access works.
func cors(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "accept, content-type")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		f(w, r)
	}
}

// verboseRequest logs and inbound request body without consuming the request
func verboseRequest(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v %v", r.Method, r.URL.Path)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Print(err)
		}
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		log.Print(string(body))

		f(w, r)
	}
}

// inboundHandler builds inbound request processing stack
func inboundHandler(f http.HandlerFunc, debug bool) http.HandlerFunc {
	if debug {
		f = verboseRequest(f)
	}
	return cors(f)
}

var apiURL = flag.String("api", "https://demo.volkszaehler.org/middleware.php", "volkszaehler api url")
var apiTimeout = flag.Duration("timeout", 30*time.Second, "volkszaehler api request timeout")
var port = flag.Int("port", 8000, "http port")
var verbose = flag.Bool("verbose", false, "verbose logging")
var help = flag.Bool("help", false, "help")

func main() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	api := newAPI(apiURL, apiTimeout, *verbose)
	server := newServer(api)

	http.HandleFunc("/", inboundHandler(server.rootHandler, *verbose))
	http.HandleFunc("/query", inboundHandler(server.queryHandler, *verbose))
	http.HandleFunc("/search", inboundHandler(server.searchHandler, *verbose))
	http.HandleFunc("/annotations", inboundHandler(server.annotationsHandler, *verbose))

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		log.Fatal(err)
	}
}
