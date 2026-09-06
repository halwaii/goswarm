package bencode

import (
	"fmt"
	"strconv"
)

// i<number>e 			 -> integer int64
// 5:hello -> 0-9 number -> string  []byte
// l<[]any>e 			 -> list	[]any
// d<key value>e -> make[string]any{ -> dictonary
// 					key : value
//					}

type Parser struct{
	data []byte // bencoded input
	pos int		// current position of data
}

func Decode(data []byte) (any, error){
	parser := Parser{
		data: data,
		pos: 0,
	}
	return parser.parse()
}

// func parse(p *Parser)
// parse() is a method that belongs to Parser
func (p *Parser) parse() (any, error){

	// bound check
	if p.pos >= len(p.data){
		return nil, fmt.Errorf("unexpected end of input")
	}

	// check
	switch p.data[p.pos]{
	case 'i':
		return p.parseInteger()

	//case 'l':
		//return parseList()
	
	//case 'd':
		//return parseDictionary()

	//case '0','1','2','3','4','5','6','7','8','9':
		//return parseString()

	default:
		return nil, fmt.Errorf("invalid bencode character : %c",p.data[p.pos])
	}
}

// integer parser
func (p *Parser) parseInteger() (int, error){
	// move position
	p.pos++

	start := p.pos
	// position should not exceed lenght of data
	for p.pos < len(p.data) && p.data[p.pos] != 'e'{
		p.pos++
	}

	if p.pos >= len(p.data){
		return 0, fmt.Errorf("unterminated integer")
	}

	// slice -> include start and exclude p.pos
	// ex: i42e
	// p.pos -> 'e', start = 4
	// numBytes = []byte{4,2} = [52 50]
	numBytes := p.data[start:p.pos]

	// string(numBytes) -> "42"
	// atoi -> ascii to integer , num = 42
	num, err := strconv.Atoi(string(numBytes))
	if err != nil{
		return 0, fmt.Errorf("invalid")
	}
	p.pos++

	return num, nil
}