package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindPrecisionFromTickSize(t *testing.T) {
	assert.Nil(t, FindPrecisionFromTickSize("1.00001"))
	assert.Nil(t, FindPrecisionFromTickSize("0000001"))
	assert.Equal(t, 5, *FindPrecisionFromTickSize("0.00001"))
	assert.Equal(t, 2, *FindPrecisionFromTickSize("0.01000000"))
}
