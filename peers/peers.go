package peers

import (
	"encoding/binary"
	"fmt"
	"net"
)

// peer structure
// peer -> 4 bytes IP + 2 bytes port = 6 bytes
// to make compact IPv4 peer list
type Peer struct {
	IP net.IP // its byte slice []byte
	Port uint16 // port range is 0-65535, so we can use uint16
}

// convert raw byte slice of peers into slice of Peer struct
func UnmarshalPeers(data []byte) ([]Peer, error){
	// each peer is 6 bytes
	if len(data)%6 != 0{
		return nil, fmt.Errorf("invalid peers data length")
	}

	numPeers := len(data)/6
	peersList := make([]Peer, numPeers)

	for i:=0;i<numPeers;i++{
		offset := i*6
		peersList[i].IP = net.IP(data[offset:offset+4])
		// interprets 2 bytes as big endian format
		peersList[i].Port = binary.BigEndian.Uint16(data[offset+4:offset+6])
	}
	return peersList, nil
}


// big Endian -> method to store multibyte data
// network protocols use -> network byte order (big endian format)