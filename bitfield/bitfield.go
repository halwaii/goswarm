package bitfield

type Bitfield []byte

// pieces are represented as byte
// 1 byte = 8 pieces

// checks if the peer has that particular piece
func HasPiece(bf *Bitfield, idx int) bool{

	// finding piece exact position
	byteIdx := idx/8
	offset := idx%8

	// edge case if piece is outside bitfield
	if byteIdx>= len(*bf){
		return false
	}

	// piece : 0 1 2 3 4 5 6 7 8
	// bits :  piece 0
	// but in Go : 7 6 5 4 3 2 1 0
	// so we do 7 - offset
	return (*bf)[byteIdx] & (1<<(7-offset)) != 0
}

// set piece marks a piece available
func SetPiece(bf *Bitfield, idx int){

	byteIdx := idx/8
	offset := idx%8

	if byteIdx>=len(*bf){
		return
	}

	(*bf)[byteIdx] |= (1<<(7-offset))
}