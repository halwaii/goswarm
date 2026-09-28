package client

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/halwaii/goswarm/bitfield"
	"github.com/halwaii/goswarm/handshake"
	"github.com/halwaii/goswarm/message"
	"github.com/halwaii/goswarm/peers"
)

// interested -> peer
//     |
// bitfield   <- peer
// 	   |
// peer tells us : i have x,y,z pieces
// state updates
// pper sends Unchoke -> Request -> peer sends pieces -> verify -> save piece
type PeerState struct { // initially
	Interested bool 	// false 
	PeerInterested bool // false
	PeerChoking bool 	// true
	Bitfield bitfield.Bitfield
}

// overall flow ****
// tcp connection -> bitTorrent handshake -> send Interested -> 
// receive bitfield -> wait for UnChoke -> send Reqeuest -> recieve piece

// request message
// request payload : piece index (4 bytes) | offset (4 bytes) | length (4 bytes)
func NewRequestMessage(idx, offset, length uint32) *message.Message{
	payload := make([]byte, 12)

	binary.BigEndian.PutUint32(payload[0:4], idx)
	binary.BigEndian.PutUint32(payload[4:8], offset)
	binary.BigEndian.PutUint32(payload[8:12], length)

	return &message.Message{
		ID: message.MsgRequest,
		Payload: payload,
	}
}

// connects to peer -> performs handshake , exchanges messages and request pieces
func ConnectToPeer (peer peers.Peer, infoHash [20]byte, peerID [20]byte) error{

	// 1) tcp connection
	peerAdd := fmt.Sprintf("%s:%d", peer.IP.String(), peer.Port)
	fmt.Printf("connecting to %s...\n", peerAdd)

	conn, err := net.DialTimeout("tcp", peerAdd, 5*time.Second)

	if err!=nil{
		return fmt.Errorf("failed tcp connection : %w", err)
	}
	defer conn.Close()
	println("TCP connection established.")

	// 2) perform bittorrent handshake
	conn.SetDeadline(time.Now().Add(5*time.Second))

	hs, err := handshake.PerformHandshake(conn, infoHash, peerID)
	if err!=nil{
		return fmt.Errorf("handshake failed : %w\n", err)
	}

	fmt.Printf("BitTorrent handshake sucessful. Peer ID : %s\n", string(hs.PeerID[:]))

	// 3) peer state initially

	state := PeerState{
		PeerChoking: true,
		Interested: false,
		PeerInterested: false,
	}
	fmt.Println("peer choking : ", state.PeerChoking)

	// 4) send interested message
	interestedMsg := &message.Message{ID: message.MsgInterested}
	_, err = conn.Write(message.Serialize(interestedMsg))
	if err != nil {
		return fmt.Errorf("failed to send interested message : %v\n", err)
	}
	state.Interested = true
	fmt.Println("interested message sent")

	// 5) wait for bitfield and other messages
	conn.SetDeadline(time.Now().Add(15*time.Second))

	for state.PeerChoking{
		msg, err := message.ReadMessage(conn)
		if err!=nil{
			return fmt.Errorf("failed to read message : %v\n", err)
		}
		if msg == nil{
			fmt.Println("received keep alive")
			continue
		}
		fmt.Println("received : ", message.String(msg))

		switch msg.ID{
		case message.MsgBitfield:
			fmt.Println("bitfield received")

			state.Bitfield = bitfield.Bitfield(msg.Payload)
			fmt.Printf("bitfield length : %d bytes\n", len(state.Bitfield))

		// have
		case message.MsgHave:
			if len(msg.Payload)!=4{
				fmt.Println("invalid HAVE message")
				continue
			}
			pieceIdx := binary.BigEndian.Uint32(msg.Payload)
			fmt.Printf("peer has piece : %d\n", pieceIdx)

			// update bitfield
			if int(pieceIdx)/8 < len(state.Bitfield){
				state.Bitfield.SetPiece(int(pieceIdx))
			}
		// choke
		case message.MsgChoke:
			state.PeerChoking=true
			fmt.Println("peer is choking us")

		// unchoke
		case message.MsgUnchoke:
			state.PeerChoking=false
			fmt.Println("peer unchoked us")
		}
	}

	// find piece
	pieceIdx := -1
	for i:=0;i<len(state.Bitfield)*8; i++{
		if state.Bitfield.HasPiece(i){
			pieceIdx=i
			break
		}
	}
	if pieceIdx == -1{
		return fmt.Errorf("peer does not have availabe piece")
	}
	fmt.Printf("selected piece : %d\n", pieceIdx)

	// send request
	const blockSize = 16*1024
	req := NewRequestMessage(uint32(pieceIdx),0, blockSize)
	
	_,err = conn.Write(message.Serialize(req))
	if err!=nil{
		return fmt.Errorf("failed to send request : %v", err)
	}
	fmt.Printf("reqeusted piece %d, offset 0, length %d\n",pieceIdx,blockSize)

	// wiat for piece
	conn.SetDeadline(time.Now().Add(15*time.Second))

	for{
		msg, err := message.ReadMessage(conn)
		if err!=nil{
			return fmt.Errorf("failed to read piece : %v", err)
		}
		if msg==nil{
			continue
		}

		if msg.ID == message.MsgPiece{
			fmt.Println("piece received")

			if len(msg.Payload)< 8{
				return fmt.Errorf("invalid piece message")
			}

			idx := binary.BigEndian.Uint32(msg.Payload[0:4])
			begin := binary.BigEndian.Uint32(msg.Payload[4:8])
			data := msg.Payload[8:]

			fmt.Printf("block index : %d\n", idx)
			fmt.Printf("block offset : %d\n", begin)
			fmt.Printf("block size bytes : %d\n", len(data))

			break
	
		}
	}
	fmt.Println("peer communication done.")
	return nil
}

