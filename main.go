package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"
)

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

	http.HandleFunc("/", handler(server.rootHandler, *verbose))
	http.HandleFunc("/query", handler(server.queryHandler, *verbose))
	http.HandleFunc("/search", handler(server.searchHandler, *verbose))
	http.HandleFunc("/annotations", handler(server.annotationsHandler, *verbose))
	http.HandleFunc("/tag-keys", handler(server.tagKeysHandler, *verbose))
	http.HandleFunc("/tag-values", handler(server.tagValuesHandler, *verbose))

	if err := http.ListenAndServe(*url, nil); err != nil {
		log.Fatal(err)
	}
}
