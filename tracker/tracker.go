package tracker

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/halwaii/goswarm/bencode"
	"github.com/halwaii/goswarm/peers"
	"github.com/halwaii/goswarm/torrent"
)

// tracerResponse -> interval
type TrackerResponse struct{
	Interval int64
	Peers []peers.Peer
}

// main function to send request to tracker
func BuildTrackerURL(t *torrent.TorrentFile,peerID [20]byte, port uint16) (string, error){

	// url.parse() -> converts string to url
	// url object -> scheme + host + path
	base, err := url.Parse(t.Announce)
	if err != nil{
		return "",err
	}

	// query parameters
	// url.values == type values map[string][]string
	// cause in http url's can have multiple values of single key
	// convert everything to string cause url accepts strings only
	// and then encode
	params := url.Values{
		"info_hash": []string{string(t.InfoHash[:])},
		"peer_id": []string{string(peerID[:])},
		"port": []string{strconv.Itoa(int(port))},
		"uploaded": []string{"0"},
		"downloaded": []string{"0"},
		"compact": []string{"1"},
		"left":[]string{strconv.FormatInt(t.Length,10)},
		"numwant":[]string{"50"},
	}
	// encode url
	base.RawQuery = params.Encode()
	// converts whole url into normal string
	// which will be sent as http request
	return base.String(),nil
}

func GeneratePeerID() ([20]byte, error){

	var peerID [20]byte

	// peerID[:] -> converts array to slice
	// cause Read requires slice of bytes
	_, err := rand.Read(peerID[:])
	if err!=nil{
		return peerID, err
	}

	return peerID, nil
}

func GetTrackerResponse(trackerURL string) ([]byte, error){
	// sent http request . resp -> *http.response
	resp, err := http.Get(trackerURL)
	if err!=nil{
		return nil, err
	}
	// execute this just before return of funciton
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		return nil, fmt.Errorf("tracker status : %s", resp.Status)
	}

	body,err := io.ReadAll(resp.Body)
	if err!=nil{
		return nil,err
	}

	return body, nil
}


func ParseTrackerResponse(data []byte) (TrackerResponse, error){
	// decode bencode
	decoded, err := bencode.Decode(data)
	if err!=nil{
		return TrackerResponse{}, err
	}

	dict, check := decoded.(map[string]any)
	if !check{
		return TrackerResponse{}, fmt.Errorf("invalid response")
	}

	interval, check := dict["interval"].(int64)
	if !check{
		return TrackerResponse{}, fmt.Errorf("invalid interval")
	}

	peersStringBytes, check := dict["peers"].(string)
	if !check{
		return TrackerResponse{}, fmt.Errorf("invalid peers")
	}

	peerList,err := peers.UnmarshalPeers([]byte(peersStringBytes))
	if err!=nil{
		return TrackerResponse{},err
	}

	return TrackerResponse{
		Interval: interval,
		Peers: peerList,
	}, nil
}