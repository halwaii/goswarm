package message

type messageID uint8

const(
	// chokes the receiver
	MsgChoke messageID = 0
	// unchokes the receiver
	MsgUnchoke messageID = 1
	// shows interest in receiving data
	MsgInterested messageID = 2
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
type Message struct{
	ID messageID
	Payload []byte
}

