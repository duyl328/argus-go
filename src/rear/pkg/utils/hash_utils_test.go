package utils

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"time"
)

const (
	fileName = "temp_rand_1GB.bin"
	fileSize = 1 << 30 // 1GB = 2^30
	bufSize  = 4 << 20 // 4MB buffer
)

func createRandomFile(path string, size int64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, bufSize)
	var written int64

	for written < size {
		n, err := rand.Read(buf)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
		w, err := f.Write(buf[:n])
		if err != nil {
			return err
		}
		written += int64(w)
		fmt.Printf("\rWritten: %.2f%%", float64(written)*100/float64(size))
	}
	fmt.Println("\nFile generation done.")
	return nil
}

func TestMain(m *testing.M) {
	fmt.Println("Creating 1GB random file...")
	if err := createRandomFile(fileName, fileSize); err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	fmt.Println("Testing SHA-256...")
	for i := 0; i < 5; i++ {
		start256 := time.Now()
		sha256Time, err := HashUtils.SHA256File(fileName)
		elapsed256 := time.Since(start256)

		if err != nil {
			fmt.Println("SHA-256 error:", err)
			return
		}
		fmt.Printf("SHA-256: %v, Speed: %.2f MB/s use time: %v\n", sha256Time, float64(fileSize)/1024/1024/elapsed256.Seconds(), elapsed256)
	}

	fmt.Println("Testing SHA-512...")
	for i := 0; i < 5; i++ {
		start512 := time.Now()
		sha512Time, err := HashUtils.SHA512File(fileName)
		elapsed512 := time.Since(start512)
		if err != nil {
			fmt.Println("SHA-512 error:", err)
			return
		}
		fmt.Printf("SHA-512: %v, Speed: %.2f MB/s use time: %v\n", sha512Time, float64(fileSize)/1024/1024/elapsed512.Seconds(), elapsed512)
	}

	// 删除临时文件
	if err := os.Remove(fileName); err != nil {
		fmt.Println("Failed to delete test file:", err)
	} else {
		fmt.Println("Temporary file deleted.")
	}
}
