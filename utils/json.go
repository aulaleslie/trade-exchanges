package utils

import (
	"encoding/json"

	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
)

// NullJSONValue is used as workaround for Golang's case-insensitive JSON
type NullJSONValue struct{}

func (m *NullJSONValue) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

func (m *NullJSONValue) UnmarshalJSON(data []byte) error {
	return nil
}

var _ json.Marshaler = (*NullJSONValue)(nil)
var _ json.Unmarshaler = (*NullJSONValue)(nil)

type APDJSON struct {
	Value *apd.Decimal
}

var _ json.Unmarshaler = (*APDJSON)(nil)

func (num *APDJSON) UnmarshalJSON(in []byte) error {
	x, err := FromStringErr(string(in))
	if err != nil {
		return errors.Wrap(err, "invalid number")
	}
	num.Value = x
	return nil
}
