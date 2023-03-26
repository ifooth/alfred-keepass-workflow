package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetFile(t *testing.T) {
	content, err := HTTPGetFile()
	assert.NoError(t, err)
	assert.Equal(t, true, len(content.Root.Groups) >= 1)
}

func TestCheckDBUpdate(t *testing.T) {
	kee := NewKee(GetKeepassURL(cfg), GetKesspassPwd(cfg))
	ok, err := kee.CheckDBUpdate()
	assert.NoError(t, err)
	assert.True(t, ok.Before(time.Now()))

	err = kee.LoadAndCache()
	assert.NoError(t, err)
	assert.True(t, len(kee.Entries) > 0)
}
