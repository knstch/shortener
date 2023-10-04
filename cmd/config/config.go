package config

import "flag"

var BasicAddr string
var Port string

func ParseFlags() {
	flag.StringVar(&Port, "a", ":8080", "port to run server")
	flag.StringVar(&BasicAddr, "b", "http://localhost"+Port+"/", "address to run server")
	flag.Parse()
}
