package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPDJSONUnmarshal(t *testing.T) {
	type X struct {
		A APDJSON
		B *APDJSON
	}
	out := X{}
	err := json.Unmarshal([]byte(`{"A":1234,"B":5678}`), &out)
	assert.NoError(t, err)

	assert.Equal(t, "1234", out.A.Value.String())

	assert.NotNil(t, out.B)
	assert.Equal(t, "5678", out.B.Value.String())
}
