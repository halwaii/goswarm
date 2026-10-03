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

	"github.com/halwaii/goswarm/client"
	"github.com/halwaii/goswarm/p2p"
	"github.com/halwaii/goswarm/torrent"
	"github.com/halwaii/goswarm/tracker"
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

	peerID, err := tracker.GeneratePeerID()

	if err!=nil{
		log.Fatal(err)
	}
	fmt.Printf("peer ID : %x\n", peerID)
	trackerURL, err := tracker.BuildTrackerURL(t, peerID, 6881)
	if err!=nil{
		log.Fatal(err)
	}
	fmt.Println(trackerURL)

	fmt.Println()

	body, err := tracker.GetTrackerResponse(trackerURL)
	if err!=nil{
		log.Fatal(err)
	}
	fmt.Println("response lenght : ", len(body))
	fmt.Println()
	
	trackerResp, err := tracker.ParseTrackerResponse(body)
	if err!=nil{
		log.Fatal(err)
	}

	fmt.Printf("interval : %d seconds\n", trackerResp.Interval)
	fmt.Printf("found %d peers\n", len(trackerResp.Peers))

	for i, peer := range trackerResp.Peers{
		fmt.Printf("\n peer %d \n", i+1)

		// 1) tcp connection
	c, err := client.New(peer, peerID, t.InfoHash)

	if err!=nil{
		fmt.Printf("failed tcp connection : %v", err)
		continue
	}
	defer c.Conn.Close()

	// send INTERESTED
	err = c.SendInterested()
	if err!=nil{
		fmt.Printf("failed to send interested : %v", err)
		continue
	}
	fmt.Println("interested message sent")

	// wait for bitfield/have + unchoke
	pieceIdx, err := c.WaitforPiece()
	if err !=nil{
		fmt.Printf("failed to get piece %v", err)
		continue
	}
	fmt.Printf("downloading piece %d\n", pieceIdx)

	// download complete piece
	piece, err := p2p.DownloadPiece(c, pieceIdx, int(t.PieceLength))
	if err!=nil{
		fmt.Printf("failed to download piece: %v\n", err)
		continue
	}

	// to verify downloaded piece
	expectedHash := t.PieceHashes[pieceIdx]

	if p2p.VerifyPiece(piece, expectedHash){
		fmt.Printf("piece %d verified\n", pieceIdx)
	} else {
		fmt.Printf("piece %d verification failed\n", pieceIdx)
	}
	break
	}
}
