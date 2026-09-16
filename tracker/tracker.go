package tracker

import (
	"crypto/rand"
	"net"
	//"net/url"
	//"github.com/halwaii/goswarm/torrent"
)

// peer structure
// peer -> 4 bytes IP + 2 bytes port = 6 bytes
// to make compact IPv4 peer list
type Peer struct {
	IP net.IP // its byte slice []byte
	Port uint16 // port range is 0-65535, so we can use uint16
}

// tracerResponse -> interval
type TrackerResponse struct{
	Interval int64
	Peers []Peer
}

// main function to send request to tracker
// func buildTrackerURL(t *torrent.TorrentFile,peerID [20]byte, port uint16) (string, error){

// 	base, err := url.Parse(t.Announce)
// 	if err != nil{
// 		return "",err
// 	}


// }

func GeneratePeerID() ([20]byte, error){

	var peerID [20]byte

	_, err := rand.Read(peerID[:])
	if err!=nil{
		return peerID, err
	}

	return peerID, nil
}