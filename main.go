package main

import (
	"bytes"
	"flag"
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

func allowedMethods(f http.HandlerFunc, methods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, allowed := range methods {
			if r.Method == allowed {
				f(w, r)
				return
			}
		}
		http.Error(w, "Bad method; supported OPTIONS, POST", http.StatusBadRequest)
	}
}

// requestLogger logs and inbound request body without consuming the request
func requestLogger(f http.HandlerFunc, debug bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var body []byte
		if debug {
			var err error
			body, err = ioutil.ReadAll(r.Body)
			if err != nil {
				log.Print(err)
			}
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}

		f(w, r)

		duration := time.Now().Sub(start)
		log.Printf("%v %v (%dms)", r.Method, r.URL.Path, duration.Nanoseconds()/1e6)

		if debug {
			log.Printf("\n" + string(body))
		}
	}
}

// inboundHandler builds inbound request processing stack
func inboundHandler(f http.HandlerFunc, debug bool) http.HandlerFunc {
	f = requestLogger(f, debug)
	f = allowedMethods(f, http.MethodOptions, http.MethodPost)
	return cors(f)
}

var apiURL = flag.String("api", "https://demo.volkszaehler.org/middleware.php", "volkszaehler api url")
var apiTimeout = flag.Duration("timeout", 30*time.Second, "volkszaehler api request timeout")
var url = flag.String("url", "0.0.0.0:8000", "listning address")
var verbose = flag.Bool("verbose", false, "verbose logging")
var help = flag.Bool("help", false, "help")

func main() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	api := newAPI(*apiURL, apiTimeout, *verbose)
	server := newServer(api, *verbose)

	http.HandleFunc("/", inboundHandler(server.rootHandler, *verbose))
	http.HandleFunc("/query", inboundHandler(server.queryHandler, *verbose))
	http.HandleFunc("/search", inboundHandler(server.searchHandler, *verbose))
	http.HandleFunc("/annotations", inboundHandler(server.annotationsHandler, *verbose))
	http.HandleFunc("/tag-keys", inboundHandler(server.tagKeysHandler, *verbose))
	http.HandleFunc("/tag-values", inboundHandler(server.tagValuesHandler, *verbose))

	if err := http.ListenAndServe(*url, nil); err != nil {
		log.Fatal(err)
	}
}
