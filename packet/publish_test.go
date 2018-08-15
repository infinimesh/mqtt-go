package packet

import "testing"
import "github.com/stretchr/testify/assert"

func TestInterpretHeaderFlags(t *testing.T) {
	input := byte(11)
	hdr, err := interpretPublishHeaderFlags(input)
	assert.NoError(t, err)
	assert.True(t, hdr.Dup)
	assert.True(t, hdr.Retain)
	assert.Equal(t, QoSLevelAtLeastOnce, hdr.QoS, "Expected at least once")
}
