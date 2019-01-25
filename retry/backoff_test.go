package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPowBackoff(t *testing.T) {
	assert := assert.New(t)

	fixtures := []struct {
		attempt  int
		expected int
	}{
		{1, 2},
		{2, 4},
		{3, 8},
		{4, 16},
		{5, 32},
	}

	for _, fixture := range fixtures {
		assert.Equal(time.Duration(fixture.expected)*time.Second, PowBackoff(fixture.attempt))
	}
}
