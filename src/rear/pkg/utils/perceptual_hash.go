package utils

import (
	"image"
	"image/color"
	"math"
	"sort"
)

type PerceptualHashType int

const (
	AHash PerceptualHashType = iota
	DHash
	PHash
)

type PerceptualHash struct {
	Value uint64
	Type  PerceptualHashType
}

type SimilarResult struct {
	Hash1      PerceptualHash
	Hash2      PerceptualHash
	Similarity float64
}

func CalculateAHash(img image.Image) PerceptualHash {
	// 缩放到8x8
	resized := resizeImage(img, 8, 8)
	
	// 转换为灰度并计算平均值
	pixels := make([]uint8, 64)
	var sum uint64
	
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			gray := rgbaToGray(resized.At(x, y))
			pixels[y*8+x] = gray
			sum += uint64(gray)
		}
	}
	
	avg := sum / 64
	
	// 生成哈希值
	var hash uint64
	for i, pixel := range pixels {
		if uint64(pixel) >= avg {
			hash |= 1 << (63 - i)
		}
	}
	
	return PerceptualHash{Value: hash, Type: AHash}
}

func CalculateDHash(img image.Image) PerceptualHash {
	// 缩放到9x8 (比较相邻像素需要多一列)
	resized := resizeImage(img, 9, 8)
	
	// 转换为灰度
	var hash uint64
	bitIndex := 0
	
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			left := rgbaToGray(resized.At(x, y))
			right := rgbaToGray(resized.At(x+1, y))
			
			if left < right {
				hash |= 1 << (63 - bitIndex)
			}
			bitIndex++
		}
	}
	
	return PerceptualHash{Value: hash, Type: DHash}
}

func CalculatePHash(img image.Image) PerceptualHash {
	// 缩放到32x32
	resized := resizeImage(img, 32, 32)
	
	// 转换为灰度
	pixels := make([][]float64, 32)
	for i := range pixels {
		pixels[i] = make([]float64, 32)
	}
	
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			pixels[y][x] = float64(rgbaToGray(resized.At(x, y)))
		}
	}
	
	// 计算DCT
	dct := computeDCT(pixels)
	
	// 取左上角8x8区域(除了DC分量)
	values := make([]float64, 64)
	idx := 0
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if y == 0 && x == 0 {
				continue // 跳过DC分量
			}
			values[idx] = dct[y][x]
			idx++
		}
	}
	
	// 计算中位数
	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	sort.Float64s(sortedValues)
	median := sortedValues[len(sortedValues)/2]
	
	// 生成哈希值
	var hash uint64
	idx = 0
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if y == 0 && x == 0 {
				continue
			}
			if dct[y][x] > median {
				hash |= 1 << (63 - idx)
			}
			idx++
		}
	}
	
	return PerceptualHash{Value: hash, Type: PHash}
}

func CalculateSimilarity(hash1, hash2 PerceptualHash) float64 {
	if hash1.Type != hash2.Type {
		return 0.0
	}
	
	// 计算汉明距离
	xor := hash1.Value ^ hash2.Value
	distance := hammingWeight(xor)
	
	// 转换为相似度百分比
	similarity := (64.0 - float64(distance)) / 64.0 * 100.0
	return math.Max(0, similarity)
}

func FindSimilarHashes(targetHash PerceptualHash, hashes []PerceptualHash, threshold float64) []SimilarResult {
	var results []SimilarResult
	
	for _, hash := range hashes {
		if hash.Type == targetHash.Type {
			similarity := CalculateSimilarity(targetHash, hash)
			if similarity >= threshold {
				results = append(results, SimilarResult{
					Hash1:      targetHash,
					Hash2:      hash,
					Similarity: similarity,
				})
			}
		}
	}
	
	// 按相似度降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})
	
	return results
}

func FindAllSimilarPairs(hashes []PerceptualHash, threshold float64) []SimilarResult {
	var results []SimilarResult
	
	for i := 0; i < len(hashes); i++ {
		for j := i + 1; j < len(hashes); j++ {
			if hashes[i].Type == hashes[j].Type {
				similarity := CalculateSimilarity(hashes[i], hashes[j])
				if similarity >= threshold {
					results = append(results, SimilarResult{
						Hash1:      hashes[i],
						Hash2:      hashes[j],
						Similarity: similarity,
					})
				}
			}
		}
	}
	
	// 按相似度降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})
	
	return results
}

// 辅助函数

func resizeImage(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x * srcW / width
			srcY := y * srcH / height
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}
	
	return dst
}

func rgbaToGray(c color.Color) uint8 {
	r, g, b, _ := c.RGBA()
	// 使用标准亮度公式
	gray := (299*r + 587*g + 114*b + 500) / 1000
	return uint8(gray >> 8)
}

func hammingWeight(x uint64) int {
	count := 0
	for x != 0 {
		count++
		x &= x - 1
	}
	return count
}

func computeDCT(pixels [][]float64) [][]float64 {
	size := len(pixels)
	dct := make([][]float64, size)
	for i := range dct {
		dct[i] = make([]float64, size)
	}
	
	for u := 0; u < size; u++ {
		for v := 0; v < size; v++ {
			sum := 0.0
			for x := 0; x < size; x++ {
				for y := 0; y < size; y++ {
					sum += pixels[x][y] * 
						math.Cos((2*float64(x)+1)*float64(u)*math.Pi/(2*float64(size))) *
						math.Cos((2*float64(y)+1)*float64(v)*math.Pi/(2*float64(size)))
				}
			}
			
			cu := 1.0
			if u == 0 {
				cu = 1.0 / math.Sqrt(2)
			}
			cv := 1.0
			if v == 0 {
				cv = 1.0 / math.Sqrt(2)
			}
			
			dct[u][v] = (2.0 / float64(size)) * cu * cv * sum
		}
	}
	
	return dct
}