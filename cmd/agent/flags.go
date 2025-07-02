package main

import (
	"flag"
	"go-metric-svc/internal/config"
)

var flags config.AgentFlagConfig

func parseFlags() {
	flag.StringVar(&flags.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flags.PoolInterval, "p", "2", "interval for pool metrics")
	flag.StringVar(&flags.SendInterval, "r", "10", "interval for send metrics")
	flag.StringVar(&flags.UseHash, "k", "", "use hash")
	flag.IntVar(&flags.WorkerCount, "l", 5, "rate limit")
	flag.StringVar(&flags.UseCrypto, "crypto-key", "", "use crypto")
	flag.StringVar(&flags.ConfigPath, "c", "", "use config")
	flag.StringVar(&flags.GRPCAddr, "gaddr", "127.0.0.1:55051", "address and port of grpc to listen on")

	flag.Parse()
}
