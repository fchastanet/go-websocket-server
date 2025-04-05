package config

import (
	"flag"
)

var (
	Addr    = ":8080"
	DevMode = false
)

func Init() {
	addr := flag.String("addr", Addr, "http service address")
	devMode := flag.Bool("dev", DevMode, "development mode")

	flag.Parse()

	Addr = *addr
	DevMode = *devMode
}
