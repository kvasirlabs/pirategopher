package pirategopher

import (
	"crypto/rand"
	"encoding/hex"
)

// Check if a value exists on slice
func stringInSlice(search string, slice []string) bool {
	for _, v := range slice {
		if v == search {
			return true
		}
	}
	return false
}

func generateRandomHexString(size int) (string, error) {
	h := make([]byte, size)
	_, err := rand.Read(h)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h)[:size], nil
}