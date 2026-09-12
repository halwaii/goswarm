package main

import (
	"fmt"
	"log"

	"github.com/halwaii/goswarm/bencode"
)

func main() {

	value := map[string]any{
		"name":   "ubuntu",
		"length": 1000,
	}

	data, err := bencode.Encode(value)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", data)
}