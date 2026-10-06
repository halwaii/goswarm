package p2p

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/halwaii/goswarm/client"
	"github.com/halwaii/goswarm/message"
	"github.com/halwaii/goswarm/peers"
	"github.com/halwaii/goswarm/torrent"
	"github.com/halwaii/goswarm/tracker"
)

// downloadPiece()
// requestBlock()
// receiveBlock()
// verifyPiece()
// DownloadPiece()

const maxblockSize = 16384

// number of unfulfilled requests we can have at a time
const maxBackLog = 5

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

// job
type pieceWork struct{
	idx int
	hash [20]byte
	length int
}
// stores reuslt
type pieceResult struct{
	idx int
	buf []byte
}

// piece progress -> state of currently downloading piece
// what is happening to one piece on one peer
// it lets wroker maintain state while receiving messages
type pieceProgress struct{
	idx int
	client *client.Client
	buf []byte
	downloaded int
	requested int
	backlog int
}

// to verify piece with our own hash
func VerifyPiece(piece []byte, expectedHash [20]byte) bool{
	actualHash := sha1.Sum(piece)

	return actualHash == expectedHash
}

// func to download file
func DownloadToFile(tf *torrent.TorrentFile, Path string)error{
	peerID, err:= tracker.GeneratePeerID()
	if err!=nil{
		return err
	}
	trackerURL, err := tracker.BuildTrackerURL(tf,peerID, 6881)
	if err!=nil{
		return err
	}

	body,err := tracker.GetTrackerResponse(trackerURL)
	if err!=nil{
		return err
	}

	trackerResp, err:= tracker.ParseTrackerResponse(body)
	if err!=nil{
		return err
	}

	t:=Torrent{
		Peers: trackerResp.Peers,
		PeerID: peerID,
		InfoHash: tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: int(tf.PieceLength),
		Length: int(tf.Length),
		Name: tf.Name,
	}

	data, err := t.Download()
	if err!=nil{
		return err
	}

	return os.WriteFile(Path, data, 0644)
}

// handles incoming messages from peer
func (state *pieceProgress) readMessage() error{
	msg, err:= state.client.Read()

	if err!=nil{
		return err
	}
	// keep alive
	if msg==nil{
		return nil
	}

	switch msg.ID{
	case message.MsgUnchoke:
		state.client.Choked = false

	case message.MsgBitfield:
		state.client.SetBitfield(msg.Payload)

	case message.MsgHave:
		state.client.SetHave(msg.Payload)

	case message.MsgChoke:
		state.client.Choked = true

	case message.MsgPiece:
		n, err := message.ParsePiece(state.idx, state.buf, msg)
		if err!=nil{
			return fmt.Errorf("failed to parse piece : %v",err)
		}
		state.downloaded += n
		state.backlog--
	}

	return nil
}

// download one piece
func DownloadPiece(pw *pieceWork, c *client.Client)([]byte, error){
	state := pieceProgress{
		idx: pw.idx,
		client: c,
		buf: make([]byte, pw.length),
	}
	// 30 second deadline
	err := c.SetDeadline(time.Now().Add(30 *time.Second))

	if err!=nil{
		return nil, err
	}
	// remove deadline  when finished
	defer c.SetDeadline(time.Time{})

	// keep going under entire piece is downloaded
	for state.downloaded< pw.length{
		// send request
		if !c.Choked{
			for state.backlog< maxBackLog && state.requested<pw.length{
				blockSize := maxblockSize
				
				// last block edge case
				if pw.length-state.requested < blockSize{
					blockSize = pw.length-state.requested
				}

				err := c.SendRequest(pw.idx, state.requested, blockSize)
				if err!=nil{
					return nil, fmt.Errorf("failed to request block : %v",err)
				}

				state.requested += blockSize
				state.backlog++
			}
		}
		// read response
		err := state.readMessage()

		if err!=nil{
			return nil, fmt.Errorf("failed to read message : %v", err)
		}
	}
	return state.buf, nil
}

// download worker
// one worker == one peer connection
func (t *Torrent) startDownloadworker(peer peers.Peer, workQueue chan *pieceWork, results chan *pieceResult){

	// connect
	c, err := client.New(peer, t.PeerID, t.InfoHash)

	if err!=nil{
		log.Printf("failed to connect to %s : %v", peer.IP, err)
		return
	}
	defer c.Conn.Close()

	// send interested
	err = c.SendInterested()

	if err!=nil{
		log.Printf("failed to send interested to %s : %v", peer.IP ,err)
		return
	}

	// receive initial peer state
	initalState := &pieceProgress{
		client: c,
	}
	// read bitfield / have / unchoke
	for c.Bitfield==nil  && len(c.AvailablePieces) == 0{
		log.Printf("waiting for bitfield from %s", peer.IP)
		err:= initalState.readMessage()

		if err!=nil{
			log.Printf("failed to receive peer state from %s : %v", peer.IP, err)
			return
		}
	}

	// piece work loop
	for pw := range workQueue{
		// doea peer have piece ?
		// if not then put it back to wrokqueue
		if !c.HasPiece(pw.idx){
			workQueue <- pw
			continue
		}

		// if yes then download piece
		buf, err := DownloadPiece(pw, c)

		if err!=nil{
			log.Printf("failed to download piece %d from %s : %v",pw.idx, peer.IP, err)
			// put it back in workqueue
			workQueue <- pw
			return
		}
		// verify sha 1
		if !VerifyPiece(buf, pw.hash){
			log.Printf("piece %d failed sha1 verification", pw.idx)
			// retry
			workQueue <- pw
			continue
		}
		log.Printf("piece %d downloaded and verified",pw.idx)

		// tell other peers we have this piece
		err = c.SendHave(pw.idx)
		if err!=nil{
			log.Printf("failed to send have message for piece %d to %s : %v", pw.idx, peer.IP, err)
		}

		// send result to results channel
		results <- &pieceResult{
			idx: pw.idx,
			buf: buf,
		}
	}
}

// calculate piece bounds
func (t *Torrent) pieceBounds(idx int)(int, int){
	start := idx * t.PieceLength
	end := start + t.PieceLength
	// last piece can be smaller than piece length
	if end > t.Length {
		end = t.Length
	}
	return start, end
}

// calculate piece size
func (t *Torrent) pieceSize(idx int) int{
	start, end := t.pieceBounds(idx)
	return end - start
}

// download entire torrent
func (t *Torrent) Download() ([]byte, error){
	// work queue
	workQueue := make(chan *pieceWork, len(t.PieceHashes))

	// results channel
	results := make(chan *pieceResult)

	// add every pieces to workqueue
	for idx, hash:= range t.PieceHashes{
		workQueue <- &pieceWork{
			idx: idx,
			hash: hash,
			length: t.pieceSize(idx),
		}
	}

	// one worker per peer
	for _,peer := range t.Peers{
		go t.startDownloadworker(peer, workQueue, results)
	}

	// final file buffer
	buf := make([]byte, t.Length)
	donepieces := 0
	for donepieces < len(t.PieceHashes){
		res := <- results

		begin, end := t.pieceBounds(res.idx)
		copy(buf[begin:end], res.buf)
		donepieces++

		percent := float64(donepieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1 // subtract 1 for main thread
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.idx, numWorkers)
	}
	close(workQueue)
	return buf,nil
}