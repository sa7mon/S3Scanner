package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRegionsDO(t *testing.T) {
	r, err := GetRegionsDO()
	assert.Nil(t, err)
	assert.Equal(t, 11, len(r))
	assert.Contains(t, r, "nyc3")
}

func TestGetRegionsLinode(t *testing.T) {
	r, err := GetRegionsLinode()
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(r), 28)
	assert.Contains(t, r, "us-east-1")
}

func TestGetRegionsScaleway(t *testing.T) {
	r, err := GetRegionsScaleway()
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(r), 3)
	assert.Contains(t, r, "fr-par")
}
