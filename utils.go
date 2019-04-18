package pirategopher

import (
	"crypto/rand"
	"encoding/hex"
)

func generateRandomHexString(size int) (string, error) {
	h := make([]byte, size)
	_, err := rand.Read(h)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h)[:size], nil
}