package main

import (
	"log"
	"os"

	"github.com/halwaii/goswarm/p2p"
	"github.com/halwaii/goswarm/torrent"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage : goswarm <torrent-file> <output-file>")
	}

	tf, err:= torrent.Open(os.Args[1])
	if err!=nil{
		log.Fatal(err)
	}

	err = p2p.DownloadToFile(tf, os.Args[2])
	if err!=nil{
		log.Fatal(err)
	}
}