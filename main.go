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
	"net"
	"os"
	"time"

	"github.com/halwaii/goswarm/handshake"
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
		//fmt.Printf("peer %d : %s %d\n", i+1, peer.IP.String(), peer.Port)

		peerAdd := fmt.Sprintf("%s:%d", peer.IP.String(), peer.Port)
		fmt.Printf("(%d) connecting to %s...\n", i+1, peerAdd)

		conn, err := net.DialTimeout("tcp", peerAdd, 5*time.Second)
		if err!=nil{
			fmt.Printf("failed to connect : %v\n",err)
			continue
		}
		defer conn.Close()

		fmt.Println("tcp connection established")
		// read and write
		conn.SetDeadline(time.Now().Add(5*time.Second))

		hs, err := handshake.PerformHandshake(conn, t.InfoHash, peerID)
		if err!=nil{
			fmt.Printf("handshake failed : %v\n", err)
			continue
		}

		fmt.Printf("BitTorrent handshake sucessful. Peer ID : %s\n", string(hs.PeerID[:]))
	}
}