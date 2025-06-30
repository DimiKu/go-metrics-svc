package main

import (
	"flag"
	"go-metric-svc/internal/config"
)

var flags config.ServerFlagConfig

func parseFlagsToStruct() {
	flag.StringVar(&flags.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flags.StoreInterval, "i", "300", "interval for save data on disc")
	flag.StringVar(&flags.FileStoragePath, "f", "/tmp/metrics-db.json", "path to local storage")
	flag.StringVar(&flags.ConnString, "d", "", "String with conn params for connect to db")
	flag.BoolVar(&flags.NeedRestore, "r", true, "path to local storage")
	flag.StringVar(&flags.UseHash, "k", "", "use hash")
	flag.StringVar(&flags.UseCrypto, "crypto-key", "", "use ssl key")
	flag.StringVar(&flags.ConfigPath, "c", "", "use config")
	flag.StringVar(&flags.TrustedSubnet, "t", "", "trusted subnet")

	flag.Parse()
}
