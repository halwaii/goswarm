package p2p

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/halwaii/goswarm/client"
	"github.com/halwaii/goswarm/message"
	"github.com/halwaii/goswarm/peers"
)

// downloadPiece()
// requestBlock()
// receiveBlock()
// verifyPiece()
// DownloadPiece()

const blockSize = 16384

// torrent holds data required to download torrent from list of peers
type Torrent struct{
	Peers []peers.Peer
	PeerID [20]byte
	InfoHash [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length int
	Name string
}
type pieceWork struct{
	idx int
	hash [20]byte
	length int
}
type pieceResult struct{
	idx int
	buf []byte
}

func DownloadPiece(c *client.Client, pieceIdx int, pieceLength int)([]byte, error){

	// buffer for complete piece
	pieceBuf := make([]byte, pieceLength)

	// downlaod piece block by block
	for offset:=0 ; offset< pieceLength; offset+=blockSize{

		// now every block will get deadline of 15 seconds
		err := c.SetDeadline(time.Now().Add(15*time.Second))
		if err!=nil{
			return nil, fmt.Errorf("failed to set deadline %w",err)
		}
		// last block can be smaller
		length := blockSize

		if offset+length > pieceLength{
			length = pieceLength - offset
		}
		fmt.Printf("requesting piece %d, offset %d, length %d\n",pieceIdx, offset, length)

		// send request to peer
		err = c.SendRequest(pieceIdx, offset, length)
		if err!=nil{
			return nil, fmt.Errorf("failed to request block %w",err)
		}

		// wait for piece response
		for {
			msg, err := c.Read()
			if err!=nil{
				return nil, fmt.Errorf("failed to read block %w",err)
			}

			// keep message alive
			if msg==nil{
				continue
			}
			// we only want piece
			if msg.ID!=message.MsgPiece{
				continue
			}

			// parse the received block
			n, err := message.ParsePiece(pieceIdx, pieceBuf, msg)
			if err!=nil{
				return nil, fmt.Errorf("failed to parse piece %w", err)
			}
			fmt.Printf("received block at offset %d (%d bytes)\n\n",offset, n)
			break
		}
	}
	fmt.Printf("\npiece %d downloaded successfully (%d bytes)\n", pieceIdx, len(pieceBuf))

	return pieceBuf, nil
}

// to verify piece with our own hash
func VerifyPiece(piece []byte, expectedHash [20]byte) bool{
	actualHash := sha1.Sum(piece)

	return actualHash == expectedHash
}