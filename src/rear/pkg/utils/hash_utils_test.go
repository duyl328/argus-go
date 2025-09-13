package utils

import (
	"crypto/rand"
	"os"
	"testing"
	"time"
)

const (
	fileName      = "temp_rand_1GB.bin"
	smallFileName = "temp_rand_small.bin"
	fileSize      = 1 << 30 // 1GB = 2^30
	smallFileSize = 1 << 20 // 1MB = 2^20
	bufSize       = 4 << 20 // 4MB buffer
)

func createRandomFile(t *testing.T, path string, size int64) error {
	if t != nil {
		t.Helper()
	}
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
		if t != nil && written%(100<<20) == 0 { // Log every 100MB
			t.Logf("Written: %.2f%%", float64(written)*100/float64(size))
		}
	}
	if t != nil {
		t.Log("File generation done.")
	}
	return nil
}

func TestHashPerformance(t *testing.T) {
	t.Log("Creating 1GB random file...")
	if err := createRandomFile(t, fileName, fileSize); err != nil {
		t.Fatalf("Error creating file: %v", err)
	}
	defer func() {
		if err := os.Remove(fileName); err != nil {
			t.Logf("Failed to delete test file: %v", err)
		} else {
			t.Log("Temporary file deleted.")
		}
	}()

	t.Log("Testing SHA-256...")
	for i := 0; i < 5; i++ {
		start256 := time.Now()
		sha256Hash, err := HashUtils.SHA256File(fileName)
		elapsed256 := time.Since(start256)

		if err != nil {
			t.Fatalf("SHA-256 error: %v", err)
		}
		t.Logf("SHA-256 #%d: %v, Speed: %.2f MB/s, Time: %v", i+1, sha256Hash, float64(fileSize)/1024/1024/elapsed256.Seconds(), elapsed256)
	}

	t.Log("Testing SHA-512...")
	for i := 0; i < 5; i++ {
		start512 := time.Now()
		sha512Hash, err := HashUtils.SHA512File(fileName)
		elapsed512 := time.Since(start512)
		if err != nil {
			t.Fatalf("SHA-512 error: %v", err)
		}
		t.Logf("SHA-512 #%d: %v, Speed: %.2f MB/s, Time: %v", i+1, sha512Hash, float64(fileSize)/1024/1024/elapsed512.Seconds(), elapsed512)
	}
}

// TestHashUtils tests basic hash functionality with a small file
func TestHashUtils(t *testing.T) {
	t.Log("Creating small test file...")
	if err := createRandomFile(t, smallFileName, smallFileSize); err != nil {
		t.Fatalf("Error creating small file: %v", err)
	}
	defer func() {
		if err := os.Remove(smallFileName); err != nil {
			t.Logf("Failed to delete small test file: %v", err)
		}
	}()

	t.Run("SHA256", func(t *testing.T) {
		hash, err := HashUtils.SHA256File(smallFileName)
		if err != nil {
			t.Fatalf("SHA-256 error: %v", err)
		}
		if len(hash) != 64 { // SHA-256 produces 64 hex characters
			t.Errorf("Expected SHA-256 hash length 64, got %d", len(hash))
		}
		t.Logf("SHA-256 hash: %s", hash)
	})

	t.Run("SHA512", func(t *testing.T) {
		hash, err := HashUtils.SHA512File(smallFileName)
		if err != nil {
			t.Fatalf("SHA-512 error: %v", err)
		}
		if len(hash) != 128 { // SHA-512 produces 128 hex characters
			t.Errorf("Expected SHA-512 hash length 128, got %d", len(hash))
		}
		t.Logf("SHA-512 hash: %s", hash)
	})

	t.Run("Consistency", func(t *testing.T) {
		// Test that the same file produces the same hash
		hash1, err := HashUtils.SHA256File(smallFileName)
		if err != nil {
			t.Fatalf("First SHA-256 error: %v", err)
		}
		hash2, err := HashUtils.SHA256File(smallFileName)
		if err != nil {
			t.Fatalf("Second SHA-256 error: %v", err)
		}
		if hash1 != hash2 {
			t.Errorf("Hash consistency failed: %s != %s", hash1, hash2)
		}
	})
}

func BenchmarkSHA256File(b *testing.B) {
	// Use smaller file for benchmarks to avoid long test times
	if err := createRandomFile(nil, smallFileName, smallFileSize); err != nil {
		b.Fatalf("Error creating file: %v", err)
	}
	defer os.Remove(smallFileName)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashUtils.SHA256File(smallFileName)
		if err != nil {
			b.Fatalf("SHA-256 error: %v", err)
		}
	}
}

func BenchmarkSHA512File(b *testing.B) {
	// Use smaller file for benchmarks to avoid long test times
	if err := createRandomFile(nil, smallFileName, smallFileSize); err != nil {
		b.Fatalf("Error creating file: %v", err)
	}
	defer os.Remove(smallFileName)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashUtils.SHA512File(smallFileName)
		if err != nil {
			b.Fatalf("SHA-512 error: %v", err)
		}
	}
}

// TestHashUtilsErrors tests error handling
func TestHashUtilsErrors(t *testing.T) {
	t.Run("NonExistentFile", func(t *testing.T) {
		_, err := HashUtils.SHA256File("non_existent_file.bin")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
		t.Logf("Expected error: %v", err)
	})

	t.Run("EmptyPath", func(t *testing.T) {
		_, err := HashUtils.SHA256File("")
		if err == nil {
			t.Error("Expected error for empty path, got nil")
		}
		t.Logf("Expected error: %v", err)
	})
}
