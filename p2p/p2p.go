package p2p

import "github.com/halwaii/goswarm/peers"

type Torrent struct {
	Peers       []peers.Peer
	InfoHash    [20]byte
	PeerID      [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}