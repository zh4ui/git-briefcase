package main

import (
	"flag"
	"log"
	"path/filepath"
)

var (
	httpAddr    = flag.String("http", ":9899", "http service address") // b=98, c=99
	templateDir = flag.String("templates", "", "a directory to look for templates and other web resources")
)

func handleFlags() {

	flag.Parse()

	if *templateDir != "" {
		if abspath, err := filepath.Abs(*templateDir); err != nil {
			log.Fatal(err)
		} else {
			*templateDir = abspath
			log.Printf("Using templateDir: %s\n", abspath)
		}
	}
}

func main() {
	handleFlags()
	server := NewGitDocityServer()
	server.Run(*httpAddr, *templateDir)
}
