package binance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderFieldsEqStringNumbers(t *testing.T) {
	of := &orderFields{}
	assert.True(t, of.equalStringNumber("0", "0"))
	assert.False(t, of.equalStringNumber("0", "a"))
	assert.False(t, of.equalStringNumber("b", "0"))

	assert.True(t, of.equalStringNumber("000123", "0123"))
	assert.True(t, of.equalStringNumber("000123.", "0123"))
	assert.True(t, of.equalStringNumber("000123.00", "0123"))
	assert.True(t, of.equalStringNumber("000123.456000", "00123.4560"))

	assert.False(t, of.equalStringNumber("123", "432"))
}
