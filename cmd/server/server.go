package main

import (
	"flag"
	"github.com/spitfire55/pirategopher"
)

var (
	flagPort = flag.Int("port", 8080, "The port to listen on")
	flagPrivateKey = flag.String("key", "private.pem", "The name of the private key file.")
	flagDb = flag.String("db", "database.db", "The location of the BoltDB database.")
)

func main () {
	flag.Parse()

	pirategopher.CreateServer(*flagPort, *flagPrivateKey, *flagDb)
}
