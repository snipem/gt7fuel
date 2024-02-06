package lib

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoundUpAlways(t *testing.T) {
	assert.Equal(t, int32(3), RoundUpAlways(2.4))
	assert.Equal(t, int32(3), RoundUpAlways(2.6))
	assert.Equal(t, int32(2), RoundUpAlways(2.0))
	assert.Equal(t, int32(3), RoundUpAlways(2.01))
	// higher is desired
	assert.Equal(t, int32(-2), RoundUpAlways(-2.01))
}
