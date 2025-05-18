package main

import "flag"

var flagRunAddr string
var storeInterval string
var fileStoragePath string
var needRestore bool
var useHash string
var connString string

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&storeInterval, "i", "300", "interval for save data on disc")
	flag.StringVar(&fileStoragePath, "f", "/tmp/metrics-db.json", "path to local storage")
	flag.StringVar(&connString, "d", "", "String with conn params for connect to db")
	flag.BoolVar(&needRestore, "r", true, "path to local storage")
	flag.StringVar(&useHash, "k", "", "use hash")

	flag.Parse()
}
