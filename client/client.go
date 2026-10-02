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
type Client struct { // initially
	Conn net.Conn
	choked bool 	// false 
	
	Bitfield bitfield.Bitfield

	Peer peers.Peer
	InfoHash [20]byte
	PeerID [20]byte

	// used when peers send HAVE message
	AvailablePieces map[int]bool
}

// connects with peer, completes handshake
func New(peer peers.Peer, peerID [20]byte, infohash [20]byte) (*Client, error){

	peerAdd := fmt.Sprintf("%s:%d", peer.IP.String(), peer.Port)

	fmt.Printf("connecting to %s\n", peerAdd)
	// perform tcp handshake
	conn,err := net.DialTimeout("tcp",peerAdd, 3*time.Second)
	if err!=nil{
		return nil, fmt.Errorf("failed to connect to peer : %w", err)
	}

	// perfor bittorrent handshake
	_, err = handshake.PerformHandshake(conn, infohash, peerID)
	if err!=nil{
		conn.Close()
		return nil, err
	}
	fmt.Println("BitTorrent handshake successful")

	return &Client{
		Conn: conn,
		choked: true,
		Peer: peer,
		PeerID: peerID,
		InfoHash: infohash,
		AvailablePieces: make(map[int]bool),
		Bitfield: nil,
	}, nil
}

// read reads and takes message from connection
func (c *Client) Read()(*message.Message, error){
	// so we can use c.Read()
	msg, err := message.ReadMessage(c.Conn)
	return msg,err
}

// send interested message
func (c *Client) SendInterested() error{
	msg := &message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(message.Serialize(msg))

	return err
}

// send request message to peer
func (c *Client) SendRequest(idx, begin, length int) error{
	req := message.MakeRequest(uint32(idx), uint32(begin), uint32(length))
	_, err := c.Conn.Write(message.Serialize(req))

	return err
}

// piece availability helper function
// now we can get info irrespective where it came from 
// that is bitfield or have
func (c *Client) HasPiece(idx int) bool{
	if c.AvailablePieces[idx]{
		return true
	}
	return c.Bitfield.HasPiece(idx)
}
// overall flow ****
// tcp connection -> bitTorrent handshake -> send Interested -> 
// receive bitfield -> wait for UnChoke -> send Reqeuest -> recieve piece


// connects to peer -> performs handshake , exchanges messages and request pieces
func ConnectToPeer (peer peers.Peer, infoHash [20]byte, peerID [20]byte) error{

	// 1) tcp connection
	c, err := New(peer, peerID, infoHash)

	if err!=nil{
		return fmt.Errorf("failed tcp connection : %w", err)
	}
	defer c.Conn.Close()

	// send INTERESTED
	err = c.SendInterested()
	if err!=nil{
		return fmt.Errorf("failed to send interested : %w", err)
	}
	fmt.Println("interested message sent")

	// wait until peer is unchoked or peer has atleast one piece
	var pieceIdx = -1
	c.Conn.SetDeadline(time.Now().Add(15*time.Second))

	for c.choked || pieceIdx==-1{

		msg, err := c.Read()
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

			c.Bitfield = bitfield.Bitfield(msg.Payload)
			for i:=0;i<len(c.Bitfield)*8;i++{
				if c.Bitfield.HasPiece(i){
					c.AvailablePieces[i]=true
				}
			}

		// have
		case message.MsgHave:
			if len(msg.Payload)!=4{
				fmt.Println("invalid HAVE message")
				continue
			}
			idx := binary.BigEndian.Uint32(msg.Payload)
			c.AvailablePieces[int(idx)]=true
			fmt.Printf("peer has piece : %d\n", int(idx))

		// choke
		case message.MsgChoke:
			c.choked=true
			fmt.Println("peer is choking us")

		// unchoke
		case message.MsgUnchoke:
			c.choked=false
			fmt.Println("peer unchoked us")
		}
		if pieceIdx==-1{
			for idx:=range c.AvailablePieces{
				pieceIdx=idx
				break
			}
		}
	}

	// find piece whenever we receive piece avaliability
	if pieceIdx==-1{
		return fmt.Errorf("peer does not have available pieces")
	}
	const blocksize = 16*1024
	err = c.SendRequest(pieceIdx, 0, blocksize)
	if err!=nil{
		return fmt.Errorf("failed to send request : %w",err)
	}

	fmt.Printf("requested piece %d , offset 0, lenght %d\n", pieceIdx,blocksize)

	// wait for piece
	for{
		msg, err := c.Read()
		if err!=nil{
			return fmt.Errorf("failed to read piece : %v", err)
		}
		if msg==nil{
			continue
		}

		if msg.ID!= message.MsgPiece{
			fmt.Println("received : ", message.String(msg))
			continue
		}

		// buffer for one full piece
		pieceBuf := make([]byte, blocksize)

		n, err := message.ParsePiece(pieceIdx, pieceBuf, msg)
		if err!=nil{
			return fmt.Errorf("failed to parse piece : %w", err)
		}
		fmt.Println("PIECE received!")
		fmt.Printf("piece index: %d\n", pieceIdx)
		fmt.Printf("block offset: 0\n")
		fmt.Printf("block size: %d bytes\n", n)

		break
	}
	fmt.Println("peer communication done.")
	return nil
}

