package handshake

import (
	"fmt"
	"io"
	"net"
)

// bit torrent protocol is of 68 bytes
// 1 byte -> prefix string size which is 19 (protocol length)
// 19 bytes -> "BitTorrent protocol" string
// 8 bytes -> reserved bytes for extensions
// 20 bytes -> infohash
// 20 bytes -> peer id

const (
	protocolString = "BitTorrent protocol"
	handshakeSize  = 68
)

// 2 types of handshakes -> TCP 3 way handshake
// 						 -> BitTorrent handshake

type Handshake struct {
	InfoHash [20]byte
	PeerID   [20]byte
}

// making bit torrent handshake using tcp connection
func PerformHandshake(conn net.Conn, infoHash [20]byte, peerID [20]byte) (*Handshake, error){

	// 1) build handshake

	buf := make([]byte, handshakeSize)
	// pstrlen
	buf[0] = byte(len(protocolString)) 
	// protocol string
	copy(buf[1:20], []byte(protocolString))
	// 8 reserved bytes are already 0
	// infoHash and peer id
	copy(buf[28:48], infoHash[:])
	copy(buf[49:68], peerID[:])

	// 2) send handshake to peer

	_, err := conn.Write(buf)
	if err!=nil{
		return nil, fmt.Errorf("failed to send handshake : %w",err)
	}

	// 3) read peer's handshake
	resp := make([]byte, handshakeSize)

	// readfull to read all bytes in one go to prevent error
	_,err = io.ReadFull(conn, resp)
	if err!=nil{
		return nil, fmt.Errorf("failed to read handshake : %w", err)
	}

	// 4) validate peer's handshake response

	if resp[0]!=byte(len(protocolString)){
		return nil, fmt.Errorf("invalied protocol length")
	}
	if string(resp[1:20]) != string(protocolString){
		return nil, fmt.Errorf("invalid protocol string")
	}

	// check if its about same torrent
	if string(resp[28:48]) != string(infoHash[:]){
		return nil, fmt.Errorf("incorrect info hash")
	}

	// 5) extract peer information
	var remoteInfoHash [20]byte
	var remotePeerID [20]byte

	copy(remoteInfoHash[:], resp[28:48])
	copy(remotePeerID[:], resp[49:68])

	return &Handshake{
		InfoHash: remoteInfoHash,
		PeerID: remotePeerID,
	}, nil
}