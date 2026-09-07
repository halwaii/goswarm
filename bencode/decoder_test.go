package bencode

import "testing"

func TestDecode(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"i42e"},
		{"i-10e"},
		{"4:spam"},
		{"0:"},
		{"l4:spam4:eggse"},
		{"li1ei2ei3ee"},
		{"lli1ei2eee"},
		{"d3:cow3:mooe"},
		{"d3:cow3:moo4:spam4:eggse"},
		{"d4:name4:John3:agei21ee"},
		{"d4:name4:John5:itemsli1ei2ei3eee"},
		{"d4:name4:John4:agei21e5:skillsl4:Go3:C++4:Rustee"},
		{"ld4:name4:John5:scorei999eei42ee"},
	}

	for _, test := range tests {
		result, err := Decode([]byte(test.input))

		if err != nil {
			t.Errorf(
				"input %q returned error: %v",
				test.input,
				err,
			)
			continue
		}

		t.Logf(
			"input: %q -> result: %#v",
			test.input,
			result,
		)
	}
}