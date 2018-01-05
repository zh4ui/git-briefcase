package main

import (
	"flag"
	"log"
	"path/filepath"
)

var (
	httpAddr  = flag.String("http", ":9899", "http service address") // b=98, c=99
	staticDir = flag.String("static-dir", "", "a directory to look for templates and other web resources")
)

func handleFlags() {

	flag.Parse()

	if *staticDir != "" {
		abs, err := filepath.Abs(*staticDir)
		if err != nil {
			log.Fatal(err)
		}

		*staticDir = abs
		log.Printf("Using staticDir: %s\n", *staticDir)
	}
}

func main() {
	handleFlags()
	server := NewDocityServer(*staticDir)
	server.Run(*httpAddr)
}
