package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
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

func allowed(f http.HandlerFunc, methods ...string) http.HandlerFunc {
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

// logger logs inbound request and body without consuming the request
func logger(f http.HandlerFunc, debug bool) http.HandlerFunc {
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

// handler builds inbound request processing stack
func handler(f http.HandlerFunc, debug bool) http.HandlerFunc {
	f = logger(f, debug)
	f = allowed(f, http.MethodOptions, http.MethodPost)
	return cors(f)
}
