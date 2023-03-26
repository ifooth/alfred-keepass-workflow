package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeeLoad(t *testing.T) {
	kee := NewKee(GetKeepassURL(cfg), GetKesspassPwd(cfg))
	err := kee.CheckDBUpdate()
	assert.NoError(t, err)
	assert.True(t, kee.needReload)

	err = kee.LoadAndCache()
	assert.NoError(t, err)
	assert.True(t, len(kee.Entries) > 0)
}
