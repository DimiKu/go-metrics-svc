package main

import "flag"

var flagRunAddr string
var storeInterval string
var fileStoragePath string
var needRestore bool

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&storeInterval, "i", "300", "interval for save data on disc")
	flag.StringVar(&fileStoragePath, "f", "/tmp/metrics-db.json", "path to local storage")
	flag.BoolVar(&needRestore, "r", true, "path to local storage")

	flag.Parse()
}
