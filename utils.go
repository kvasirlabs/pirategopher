package pirategopher

import (
	"crypto/rand"
	"encoding/hex"
	"os"
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

func getDrives() (drives []string) {
	for _, letter := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		drive := string(letter) + ":\\"
		_, err := os.Stat(drive)
		if err == nil {
			drives = append(drives, drive)
		}
	}
	return drives
}