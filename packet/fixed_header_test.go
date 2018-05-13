package packet

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRemainingLength(t *testing.T) {

	var testCases = []struct {
		input    []byte
		expected int
	}{
		{
			input:    []byte{byte(20)},
			expected: 20,
		}, {
			input:    []byte{byte(64)},
			expected: 64,
		},
		{
			input:    []byte{byte(193), byte(2)},
			expected: 321,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			input := bytes.NewBuffer(tc.input)
			actual, err := getRemainingLength(input)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}

}

// TODO test what heppens if it says more packets, but none comes.. ?
