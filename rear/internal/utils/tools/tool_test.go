package tools

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ========== 使用示例 ==========

func Example() {
	ctx := context.Background()

	// 1. 简单的图片处理
	err := ResizeImage(ctx, "input.jpg", "output.jpg", 800, 600)
	if err != nil {
		fmt.Printf("Resize failed: %v\n", err)
	}

	// 2. 获取图片信息
	//info, err := GetImageInfo(ctx, "photo.jpg")
	//if err != nil {
	//	fmt.Printf("Get info failed: %v\n", err)
	//}
	//fmt.Printf("Image: %dx%d, Format: %s\n", info.Width, info.Height, info.Format)

	// 3. 处理 EXIF
	exifData, err := GetExifData(ctx, "photo.jpg")
	if err != nil {
		fmt.Printf("Get EXIF failed: %v\n", err)
	}
	fmt.Printf("EXIF: %v\n", exifData)

	// 4. 在 goroutine 中使用（并发安全）
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// 每个 goroutine 可以安全地调用这些函数
			input := fmt.Sprintf("img%d.jpg", index)
			output := fmt.Sprintf("thumb%d.jpg", index)

			// 使用带超时的 context
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := ResizeImage(ctx, input, output, 200, 200); err != nil {
				fmt.Printf("Error processing %s: %v\n", input, err)
			}
		}(i)
	}
	wg.Wait()
}
