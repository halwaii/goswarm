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

	case 'l':
		return p.parseList()
	
	case 'd':
		return p.parseDictionary()

	case '0','1','2','3','4','5','6','7','8','9':
		return p.parseString()

	default:
		return nil, fmt.Errorf("invalid bencode character : %c",p.data[p.pos])
	}
}

// integer parser
func (p *Parser) parseInteger() (int64, error){
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
	//num, err := strconv.Atoi(string(numBytes))

	// used parseInt to make num 64-bit
	num, err := strconv.ParseInt(string(numBytes), 10, 64)
	if err != nil{
		return 0, fmt.Errorf("invalid integer '%s': %v", string(numBytes), err)
	}
	p.pos++

	return num, nil
}

// string parser
func (p *Parser) parseString() (string, error){
	start := p.pos

	// find " : "
	for p.pos<len(p.data) && p.data[p.pos]!=':'{
		p.pos++
	}

	// if not foudn
	if p.pos>=len(p.data){
		return "", fmt.Errorf("invalid string length")
	}

	// find length -> from start to p.pos-1
	bytesLength := p.data[start:p.pos]
	length, err := strconv.Atoi(string(bytesLength))
	if err!=nil{
		return "",fmt.Errorf("invalid length")
	}

	// skip ':'
	p.pos++

	// check bytesLength and data lenght
	if p.pos+length > len(p.data){
		return "",fmt.Errorf("error")
	}

	// extract actual string
	val := string(p.data[p.pos:p.pos+length])

	// move on
	p.pos += length

	return val, nil
}

// list parser
func (p *Parser) parseList() ([]any, error){
	// skip 'l'
	p.pos++

	// list can have anything -> int, string, list, dictionary
	list := []any{}

	for {
		// check bounds
		if p.pos>=len(p.data){
			return nil, fmt.Errorf("unterminated list")
		}

		// list ends when 'e'
		if p.data[p.pos] == 'e'{
			p.pos++
			break
		}

		// parse next element
		// recursive parsing
		value, err := p.parse()
		if err!= nil{
			return nil, err
		}

		// add element
		list = append(list, value)
	}
	return list,nil
}

// dictionary parser
func (p *Parser) parseDictionary() (map[string]any, error){

	// skip 'd'
	p.pos++

	dict := make(map[string]any)

	for{
		// check bounds
		if p.pos>=len(p.data){
			return nil, fmt.Errorf("unterminated dictionary")
		}

		// break condition
		if p.data[p.pos] == 'e'{
			p.pos++
			break
		}

		// dictonary -> key : value
		// keys msut be strings
		key, err := p.parseString()
		if err!=nil{
			return nil, err
		}

		// parse the corresponding value
		value, err := p.parse()
		if err!=nil{
			return nil,err
		}

		// store
		dict[key] = value
	}
	return dict,nil
}