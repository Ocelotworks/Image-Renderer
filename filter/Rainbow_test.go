package filter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlendColours(t *testing.T) {
	assert.Equal(t, uint8(255), blendColours(1, 255))

}
