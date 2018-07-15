package packet

import "testing"
import "github.com/stretchr/testify/assert"

func TestInterpretHeaderFlags(t *testing.T) {
	input := byte(11)
	dup, retain, qos, err := interpretHeaderFlags(input)
	assert.NoError(t, err)
	assert.True(t, dup)
	assert.True(t, retain)
	assert.Equal(t, qosLevelAtLeastOnce, qos, "Expected at least once")
}
