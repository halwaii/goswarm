package torrent

import (
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/halwaii/goswarm/bencode"
)

// torrent
// │
// ├── announce -> tracker's url
// │
// └── info
//      │
//      ├── name 			-> file name
//      ├── length			-> file size
//      ├── piece length	-> file's byte pieces
//      └── pieces			-> sha 1 hash of every piece

// pieces = sha1(piece0) + sha1(piece1) + sha1(piece2) + sha1(piece3)
// each sha1 is of 20 bytes

type TorrentFile struct{
	Announce string
	Name string
	Length int64
	PieceLength int64
	PieceHashes [][20]byte // slice of 20 byte arrays

	InfoHash [20]byte // sha1 hash of info dictionary
}

// torrnet parser => read file -> decode bencode -> extract values -> return torrentFile

// reads a .torrent file and convert it into torrent metadata
func Open(path string) (*torrentFile, error){

	// 1) read .torrent file
	data, err := os.ReadFile(path)
	if err != nil{
		return nil, fmt.Errorf("failed to read torrent file : %w", err)
		// %w -> used with errorf to wrap an error
	}

	// 2) decode bencode
	decoded, err := bencode.Decode(data)
	if err!=nil{
		return nil, fmt.Errorf("failed to decode torrent file : %w",err)
	}

	// 3) root of torrent file must be dictionary
	root, check := decoded.(map[string]any)
	if !check{
		return nil, fmt.Errorf("root is not a dictionary")
	}

	// 4) get announce url
	// extract announce inside root which must be a string
	announce, check := root["announce"].(string)
	if !check{
		return nil, fmt.Errorf("invalid announce field")
	}

	// 5) get info dictionary
	info, check := root["info"].(map[string]any)
	if !check{
		return nil, fmt.Errorf("invalid info dictionary")
	}

	// 5.5) calculate info hash
	infoBytes, err := bencode.Encode(info)
	if err!=nil{
		return nil, fmt.Errorf("failed to encode info dictionary : %w", err)
	}

	// extract metadata from info dictionary and calculate sha1 hash of it

	infoHash := sha1.Sum(infoBytes)
	// sha1 hash of info dictionary is called info hash
	// info hash is used to identify the torrent file in the tracker

	// 6) get file name
	name, check := info["name"].(string)
	if !check{
		return nil, fmt.Errorf("invalid name")
	}

	// 7) get file length
	length, check := info["length"].(int64)
	if !check{
		return nil, fmt.Errorf("invalid length")
	}

	// 8) get piece length
	PieceLength, check := info["piece length"].(int64)
	if !check{
		return nil, fmt.Errorf("invalid piece length")
	}

	// 9) get piece hash in form of string
	piecesString, check := info["pieces"].(string)
	if !check{
		return nil, fmt.Errorf("invalid pieces")
	}
	// convert the string into byte
	pieces := []byte(piecesString)

	// 10) validate every values
	if length<0{
		return nil, fmt.Errorf("file length cannot be negative")
	}
	if PieceLength<=0{
		return nil, fmt.Errorf("piece length cannot be negative or 0")
	}
	// every sha 1 must be 20 bytes
	if len(pieces)%20 !=0{
		return nil, fmt.Errorf("pieces length is not a multiple of 20")
	}

	// convert pieces into individual [20]byte hashes
	// [][20]byte{ of that many pieces present }
	PieceHashes := make([][20]byte, len(pieces)/20)

	// slice every 20 byte chunk and copy in array
	for i:=0;i < len(PieceHashes);i++{
		
		start := i*20
		end := start+20

		// pieces[start:end] will give slice of 20 bytes
		// copy() -> duplicates elements from source slice into destination slice
		copy(PieceHashes[i][:], pieces[start:end])
	}

	// create our torrent file
	torrent := &TorrentFile{
		Announce: announce,
		Name: name,
		Length: length,
		PieceLength: PieceLength,
		PieceHashes: PieceHashes,

		InfoHash: infoHash,
	}

	return torrent, nil
}

