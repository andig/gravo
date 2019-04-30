package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/andig/gravo/volkszaehler"
)

var (
	version = "development"
	commit  = "unknown commit"
)

var apiURL = flag.String("api", "https://demo.volkszaehler.org/middleware.php", "volkszaehler api url")
var apiTimeout = flag.Duration("timeout", 30*time.Second, "volkszaehler api request timeout")
var url = flag.String("url", "0.0.0.0:8000", "listening address")
var verbose = flag.Bool("verbose", false, "verbose logging")
var help = flag.Bool("help", false, "help")

func main() {
	log.Printf("Running gravo %s (%s)", version, commit)

	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	client := volkszaehler.NewClient(*apiURL, apiTimeout, *verbose)
	server := newServer(client)

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
