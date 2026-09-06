package bencode

import "testing"

func TestParseInteger(t *testing.T) {

    tests := []struct {
        input    string
        expected int
    }{
        {"i42e", 100},
        {"i0e", 0},
        {"i-10e", -10},
        {"i999e", 999},
    }

    for _, test := range tests {

        result, err := Decode([]byte(test.input))

        if err != nil {
            t.Errorf(
                "input %s returned error: %v",
                test.input,
                err,
            )
            continue
        }

        if result != test.expected {
            t.Errorf(
                "input %s: expected %d, got %v",
                test.input,
                test.expected,
                result,
            )
        }
    }
}