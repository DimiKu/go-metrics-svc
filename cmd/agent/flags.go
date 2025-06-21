package main

import "flag"

var (
	flagRunAddr  string
	poolInterval string
	sendInterval string
	useHash      string
	workerCount  int
	useCrypto    string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&poolInterval, "p", "2", "interval for pool metrics")
	flag.StringVar(&sendInterval, "r", "10", "interval for send metrics")
	flag.StringVar(&useHash, "k", "", "use hash")
	flag.IntVar(&workerCount, "l", 5, "rate limit")
	flag.StringVar(&useCrypto, "c", "", "use crypto")

	flag.Parse()
}
