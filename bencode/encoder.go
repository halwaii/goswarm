package bencode

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
)

// decoder => bencode bytes -> decode -> value
// encoder => value -> encode -> bencode bytes

func Encode(val any) ([]byte, error) {

	// buffer of bytes -> read and write methods
	// good for step by step data building
	var buf bytes.Buffer

	err := encodeVal(&buf, val)
		if err!= nil{
			return nil,err
		}

	// this will return []byte
	return buf.Bytes(), nil
}

// recursive encoding 
func encodeVal(buf *bytes.Buffer, val any) error {
	switch v := val.(type){

	case int64:
		return encodeInteger(buf,v)

	case string:
		return encodeString(buf, v)

	// new case for binary data
	case []byte:
		return encodeByteSlice(buf, v)
	
	case []any:
		return encodeList(buf, v)

	case map[string]any:
		return encodeDictionary(buf,v)

	default:
		return fmt.Errorf("invalid : %T", val)
	}
}

// integer encoding
func encodeInteger(buf *bytes.Buffer, val int64) error{
	buf.WriteByte('i')
	// convert integer to string then convert it into string
	// convert int 64 to string base 10
	buf.WriteString(strconv.FormatInt(val,10))
	buf.WriteByte('e')

	return nil
}

// string encoding
func encodeString(buf *bytes.Buffer, val string) error{
	buf.WriteString(strconv.Itoa(len(val)))
	buf.WriteByte(':')
	buf.WriteString(val)

	return nil
}

// list encoding
// l5:helloi10ee
// ["hello", 10]
func encodeList(buf *bytes.Buffer, list []any) error{

	buf.WriteByte('l')

	for _, val := range list{
		err:= encodeVal(buf, val)
		if err!= nil{
			return err
		}
	}
	buf.WriteByte('e')

	return nil
}

// dictionary encoding
// bencode dictionary keys must be sorted lexicographically
func encodeDictionary(buf *bytes.Buffer, dict map[string]any) error{

	buf.WriteByte('d')
	// dict -> key(string) , value

	// memory is unordered in go maps
	// make an array of dictionary size
	keys := make([]string, 0, len(dict))

	// now store all keys in array
	for key := range dict{
		keys = append(keys, key)
	}

	// sort those keys
	sort.Strings(keys)

	// encode key value pair
	for _, key := range keys{
		// key is string
		err := encodeString(buf, key)
		if err!= nil{
			return err
		}

		// encode value
		err = encodeVal(buf, dict[key])
		if err!= nil{
			return err
		}
	}
	buf.WriteByte('e')

	return nil
}

// byte slice encoding for raw binary data -> torrent pieces
func encodeByteSlice(buf *bytes.Buffer, val []byte) error{
	buf.WriteString(strconv.Itoa(len(val)))
	buf.WriteByte(':')
	buf.Write(val) // store raw binary data as it is

	return nil
}