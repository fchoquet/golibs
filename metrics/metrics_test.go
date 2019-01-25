package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTags(t *testing.T) {
	assert := assert.New(t)

	c1 := &client{
		tags: []string{},
	}

	c2 := c1.WithTags([]string{"foo", "bar"})

	assert.Equal([]string{}, c1.tags)
	assert.Equal([]string{"foo", "bar"}, c2.(*client).tags)

	c3 := c2.WithTags([]string{"baz"})
	assert.Equal([]string{"foo", "bar"}, c2.(*client).tags)
	assert.Equal([]string{"foo", "bar", "baz"}, c3.(*client).tags)
}
