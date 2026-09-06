package bencode

import "fmt"

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
		parseInteger();

	case 'l':
		parseList();
	
	case 'd':
		parseDictionary();

	case '0','1','2','3','4','5','6','7','8','9':
		parseString();
	}

}