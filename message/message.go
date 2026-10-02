package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

const (
	// chokes the receiver
	MsgChoke messageID = 0
	// unchokes the receiver
	MsgUnchoke messageID = 1
	// shows interest in receiving data
	MsgInterested    messageID = 2
	MsgNotinterested messageID = 3
	// alerts the receiver that sender has downloaded a piece
	MsgHave messageID = 4
	// encodes which pieces the sender has downloaded
	MsgBitfield messageID = 5
	// requests a block of data from receiver
	MsgRequest messageID = 6
	// delivers a block of data to fulfill a request
	MsgPiece messageID = 7
	// cancels a request
	MsgCancel messageID = 8
)

// a message stores id and payload
// -------------------------------
// | 4-byte length | ID | payload |
// -------------------------------
type Message struct {
	ID      messageID
	Payload []byte
}

// request message
// request payload : piece index (4 bytes) | offset (4 bytes) | length (4 bytes)
func MakeRequest(idx, offset, length uint32) *Message{
	payload := make([]byte, 12)

	binary.BigEndian.PutUint32(payload[0:4], idx)
	binary.BigEndian.PutUint32(payload[4:8], offset)
	binary.BigEndian.PutUint32(payload[8:12], length)

	return &Message{
		ID: MsgRequest,
		Payload: payload,
	}
}

// MakeHave creates HAVE message
func MakeHave(idx int) *Message{
	// payload => piece index | offset | length
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(idx))

	return &Message{
		ID: MsgHave,
		Payload: payload,
	}
}

// parsepiece parses a piece message and copies block data into buf
func ParsePiece(idx int,buf []byte, msg *Message) (int, error){
	if msg.ID != MsgPiece{
		return 0, fmt.Errorf("expected PIECE (ID %d), got ID %d", MsgPiece, msg.ID)
	}

	// piece index 4 bytes + offset 4 bytes
	if len(msg.Payload)< 8{
		return 0, fmt.Errorf("payload too short : %d < 8", len(msg.Payload))
	}

	parsedIdx := int(binary.BigEndian.Uint32(msg.Payload[0:4]))

	if parsedIdx != idx{
		return 0, fmt.Errorf("expected index : %d, got %d", idx, parsedIdx)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))

	data := msg.Payload[8:]

	if begin>len(buf){
		return 0, fmt.Errorf("begin offset too high")
	}

	if begin + len(data)> len(buf){
		return 0, fmt.Errorf("Data too long [%d] for offset %d with length %d", len(data), begin, len(buf))
	}

	copy(buf[begin:], data)
	return len(data), nil
}

// parse a have message
func ParseHave(msg *Message) (int, error){
	if msg.ID!=MsgHave{
		return 0, fmt.Errorf("Expected HAVE (ID %d), got ID %d", MsgHave, msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("Expected payload length 4, got length %d", len(msg.Payload))
	}
	idx := int(binary.BigEndian.Uint32(msg.Payload))
	return idx,nil
}
// serialize => message struct -> byte[]
func Serialize(m *Message) []byte {

	// if nil keep message alive
	if m == nil {
		return make([]byte, 4)
	}

	// calculate length -> ID + payload
	length := uint32(len(m.Payload) + 1)

	// creating buffer ( 4 bytes -> length prefix)
	buf := make([]byte, 4+length)

	// convert uint32 into 4 bytes
	binary.BigEndian.PutUint32(buf[0:4], length)
	// write message id and payload
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)

	return buf
}

// deserialization
// tcp bytes -> readmessage -> message struct
// directly passing tc connection
func ReadMessage(r io.Reader) (*Message, error){
	// 1st 4 bytes message length
	lengthBuf := make([]byte, 4)

	// read all 4 bytes completely
	_,err := io.ReadFull(r, lengthBuf)
	if err!=nil{
		return nil, err
	}

	// decode length
	// bytes -> uint32
	length := binary.BigEndian.Uint32(lengthBuf)

	// keep message alive
	if length==0{
		return nil,nil
	}

	// read actual message
	msgBuf := make([]byte, length)
	_,err = io.ReadFull(r, msgBuf)
	if err!=nil{
		return nil, err
	}

	m := &Message{
		ID: messageID(msgBuf[0]),
		Payload: msgBuf[1:],
	}

	return m, nil

}

// to test
func String(m *Message) string{
	if m==nil{
		return "keep alive"
	}
	return fmt.Sprintf("message{ID : %d, payload : %x}", m.ID, m.Payload)
}