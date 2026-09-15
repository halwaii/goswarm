package tracker

import (
	"net"

	"github.com/halwaii/goswarm/torrent"
)

// peer structure
type Peer struct {
	IP net.IP // its byte slice []byte
	Port uint16
}

// tracerResponse -> interval
type TrackerResponse struct{
	Interval int64
	Peers []Peer
}

// main function to send request to tracker
func Announce(t *torrent.TorrentFile, peerID [20]byte, port uint16) ()