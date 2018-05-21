//--------------------------------------------------------------------------
// Copyright 2018 infinimesh, INC
// www.infinimesh.io
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//--------------------------------------------------------------------------

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
