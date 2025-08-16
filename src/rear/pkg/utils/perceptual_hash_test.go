package utils

import (
	"image"
	"image/color"
	"testing"
)

func createTestImage(width, height int, pattern func(x, y int) color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, pattern(x, y))
		}
	}
	return img
}

func TestCalculateAHash(t *testing.T) {
	// 创建一个简单的测试图像
	img := createTestImage(100, 100, func(x, y int) color.Color {
		if (x+y)%2 == 0 {
			return color.RGBA{255, 255, 255, 255} // 白色
		}
		return color.RGBA{0, 0, 0, 255} // 黑色
	})
	
	hash := CalculateAHash(img)
	
	if hash.Type != AHash {
		t.Errorf("Expected hash type %v, got %v", AHash, hash.Type)
	}
	
	if hash.Value == 0 {
		t.Error("Hash value should not be zero for a patterned image")
	}
}

func TestCalculateDHash(t *testing.T) {
	// 创建一个渐变图像
	img := createTestImage(100, 100, func(x, y int) color.Color {
		intensity := uint8(float64(x) / 100.0 * 255)
		return color.RGBA{intensity, intensity, intensity, 255}
	})
	
	hash := CalculateDHash(img)
	
	if hash.Type != DHash {
		t.Errorf("Expected hash type %v, got %v", DHash, hash.Type)
	}
	
	if hash.Value == 0 {
		t.Error("Hash value should not be zero for a gradient image")
	}
}

func TestCalculatePHash(t *testing.T) {
	// 创建一个更复杂的测试图像
	img := createTestImage(100, 100, func(x, y int) color.Color {
		intensity := uint8((x*y) % 256)
		return color.RGBA{intensity, intensity, intensity, 255}
	})
	
	hash := CalculatePHash(img)
	
	if hash.Type != PHash {
		t.Errorf("Expected hash type %v, got %v", PHash, hash.Type)
	}
}

func TestCalculateSimilarity(t *testing.T) {
	// 创建两个相似的图像
	img1 := createTestImage(100, 100, func(x, y int) color.Color {
		if x < 50 {
			return color.RGBA{255, 255, 255, 255}
		}
		return color.RGBA{0, 0, 0, 255}
	})
	
	img2 := createTestImage(100, 100, func(x, y int) color.Color {
		if x < 48 { // 稍微不同
			return color.RGBA{255, 255, 255, 255}
		}
		return color.RGBA{0, 0, 0, 255}
	})
	
	hash1 := CalculateAHash(img1)
	hash2 := CalculateAHash(img2)
	
	similarity := CalculateSimilarity(hash1, hash2)
	
	if similarity < 50.0 {
		t.Errorf("Expected similarity > 50%%, got %f%%", similarity)
	}
	
	if similarity > 100.0 {
		t.Errorf("Similarity should not exceed 100%%, got %f%%", similarity)
	}
}

func TestCalculateSimilarityDifferentTypes(t *testing.T) {
	img := createTestImage(100, 100, func(x, y int) color.Color {
		return color.RGBA{128, 128, 128, 255}
	})
	
	ahash := CalculateAHash(img)
	dhash := CalculateDHash(img)
	
	similarity := CalculateSimilarity(ahash, dhash)
	
	if similarity != 0.0 {
		t.Errorf("Expected 0%% similarity for different hash types, got %f%%", similarity)
	}
}

func TestFindSimilarHashes(t *testing.T) {
	// 创建测试图像和哈希
	img1 := createTestImage(100, 100, func(x, y int) color.Color {
		return color.RGBA{255, 255, 255, 255}
	})
	
	img2 := createTestImage(100, 100, func(x, y int) color.Color {
		return color.RGBA{250, 250, 250, 255} // 非常相似
	})
	
	img3 := createTestImage(100, 100, func(x, y int) color.Color {
		return color.RGBA{0, 0, 0, 255} // 完全不同
	})
	
	hash1 := CalculateAHash(img1)
	hash2 := CalculateAHash(img2)
	hash3 := CalculateAHash(img3)
	
	hashes := []PerceptualHash{hash2, hash3}
	results := FindSimilarHashes(hash1, hashes, 80.0)
	
	if len(results) == 0 {
		t.Error("Expected to find at least one similar hash")
	}
	
	// 检查结果是否按相似度排序
	for i := 1; i < len(results); i++ {
		if results[i-1].Similarity < results[i].Similarity {
			t.Error("Results should be sorted by similarity in descending order")
		}
	}
}

func TestFindAllSimilarPairs(t *testing.T) {
	// 创建几个测试哈希
	img1 := createTestImage(100, 100, func(x, y int) color.Color {
		return color.RGBA{255, 255, 255, 255}
	})
	
	img2 := createTestImage(100, 100, func(x, y int) color.Color {
		return color.RGBA{250, 250, 250, 255}
	})
	
	img3 := createTestImage(100, 100, func(x, y int) color.Color {
		return color.RGBA{245, 245, 245, 255}
	})
	
	hash1 := CalculateAHash(img1)
	hash2 := CalculateAHash(img2)
	hash3 := CalculateAHash(img3)
	
	hashes := []PerceptualHash{hash1, hash2, hash3}
	results := FindAllSimilarPairs(hashes, 70.0)
	
	if len(results) == 0 {
		t.Error("Expected to find similar pairs")
	}
	
	// 验证没有重复的配对
	seen := make(map[string]bool)
	for _, result := range results {
		key1 := string(rune(result.Hash1.Value)) + "-" + string(rune(result.Hash2.Value))
		key2 := string(rune(result.Hash2.Value)) + "-" + string(rune(result.Hash1.Value))
		
		if seen[key1] || seen[key2] {
			t.Error("Found duplicate pair in results")
		}
		seen[key1] = true
		seen[key2] = true
	}
}

func TestHammingWeight(t *testing.T) {
	tests := []struct {
		input    uint64
		expected int
	}{
		{0, 0},
		{1, 1},
		{3, 2},  // 11 in binary
		{7, 3},  // 111 in binary
		{15, 4}, // 1111 in binary
		{0xFFFFFFFFFFFFFFFF, 64}, // all bits set
	}
	
	for _, test := range tests {
		result := hammingWeight(test.input)
		if result != test.expected {
			t.Errorf("hammingWeight(%d) = %d; expected %d", test.input, result, test.expected)
		}
	}
}

func TestRgbaToGray(t *testing.T) {
	tests := []struct {
		input    color.RGBA
		expected uint8
	}{
		{color.RGBA{255, 255, 255, 255}, 255}, // 白色
		{color.RGBA{0, 0, 0, 255}, 0},         // 黑色
		{color.RGBA{255, 0, 0, 255}, 76},      // 红色
		{color.RGBA{0, 255, 0, 255}, 149},     // 绿色
		{color.RGBA{0, 0, 255, 255}, 29},      // 蓝色
	}
	
	for _, test := range tests {
		result := rgbaToGray(test.input)
		// 允许一定的误差
		if result < test.expected-2 || result > test.expected+2 {
			t.Errorf("rgbaToGray(%v) = %d; expected around %d", test.input, result, test.expected)
		}
	}
}

func BenchmarkCalculateAHash(b *testing.B) {
	img := createTestImage(1000, 1000, func(x, y int) color.Color {
		return color.RGBA{uint8(x % 256), uint8(y % 256), uint8((x + y) % 256), 255}
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateAHash(img)
	}
}

func BenchmarkCalculateDHash(b *testing.B) {
	img := createTestImage(1000, 1000, func(x, y int) color.Color {
		return color.RGBA{uint8(x % 256), uint8(y % 256), uint8((x + y) % 256), 255}
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateDHash(img)
	}
}

func BenchmarkCalculatePHash(b *testing.B) {
	img := createTestImage(1000, 1000, func(x, y int) color.Color {
		return color.RGBA{uint8(x % 256), uint8(y % 256), uint8((x + y) % 256), 255}
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculatePHash(img)
	}
}