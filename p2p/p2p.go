package p2p

import (
	"fmt"

	"github.com/halwaii/goswarm/client"
	"github.com/halwaii/goswarm/message"
)

// downloadPiece()
// requestBlock()
// receiveBlock()
// verifyPiece()
// DownloadPiece()

const blockSize = 16384

func DownloadPiece(c *client.Client, pieceIdx int, pieceLength int)([]byte, error){

	// buffer for complete piece
	pieceBuf := make([]byte, pieceLength)

	// downlaod piece block by block
	for offset:=0 ; offset< pieceLength; offset+=blockSize{
		// last block can be smaller
		length := blockSize

		if offset+length > pieceLength{
			length = pieceLength - offset
		}
		fmt.Printf("requesting piece %d, offset %d, length %d\n",pieceIdx, offset, length)

		// send request to peer
		err := c.SendRequest(pieceIdx, offset, length)
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