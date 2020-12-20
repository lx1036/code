package main

import (
	"flag"
	"log"
	"net/http"
)

var addr = flag.String("address", ":8080", "The address to listen on for HTTP requests.")

func main() {
	flag.Parse()
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
