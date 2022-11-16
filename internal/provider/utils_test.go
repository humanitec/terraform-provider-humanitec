package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueAtPath(t *testing.T) {
	assert := assert.New(t)

	testObj := map[string]interface{}{
		"some": map[string]interface{}{
			"nested": map[string]interface{}{
				"property": []string{"a", "b"},
			},
		},
	}

	value, ok := valueAtPath[[]string](testObj, []string{"some", "nested", "property"})
	assert.True(ok)
	assert.Equal([]string{"a", "b"}, value)
}
