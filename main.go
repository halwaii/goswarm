// package main

// import (
// 	"fmt"
// 	"log"

// 	"github.com/halwaii/goswarm/bencode"
// )

// func main() {

// 	value := map[string]any{
// 		"name":   "ubuntu",
// 		"length": 1000,
// 	}

// 	data, err := bencode.Encode(value)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("%s\n", data)
// }

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/halwaii/goswarm/torrent"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: goswarm <torrent-file>")
		return
	}

	t, err := torrent.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Torrent Information")
	fmt.Println()

	fmt.Println("Name:", t.Name)
	fmt.Println("Length:", t.Length)
	fmt.Println("Piece Length:", t.PieceLength)
	fmt.Println("Number of Pieces:", len(t.PieceHashes))
	fmt.Println("Tracker:", t.Announce)
	fmt.Printf("Info Hash: %x\n", t.InfoHash)
}