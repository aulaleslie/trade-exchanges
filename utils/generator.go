package utils

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/pkg/errors"
)

func GenClientOrderID(identifierID string) (string, error) {
	// TODO: cover this with test, maybe? Test error cases? Same parameters cases?
	// TODO: make returning same ID for same parameters of orders
	// guid, err := uuid.NewV4()
	// if err != nil {
	// 	return "", errors.Wrap(err, "can't generate UUID")
	// }

	// TODO: make order ID more uqiuely (longer)?
	hexID := [4]byte{}
	_, err := rand.Read(hexID[:])
	if err != nil {
		return "", errors.Wrap(err, "can't generate ID")
	}

	prefferedID := "RUN" + identifierID + "-" + hex.EncodeToString(hexID[:])
	return prefferedID, nil
}
