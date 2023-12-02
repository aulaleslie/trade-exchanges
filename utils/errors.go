package utils

import (
	"github.com/pkg/errors"
)

func ReplaceError(new, old error) error {
	return errors.Wrapf(new, "[Original: %v]", old)
}
