package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFile(t *testing.T) {
	content, err := HTTPGetFile()
	assert.NoError(t, err)
	assert.Equal(t, true, len(content.Root.Groups) >= 1)
}
