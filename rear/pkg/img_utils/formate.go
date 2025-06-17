package imgutils

var (
	// 常规格式
	CommonFormat = []string{"JPG", "JPEG", "PNG"}
	// 需要特殊处理的格式
	UncommonFormat = []string{"BMP", "TIFF", "TIF", "WebP", "HEIC", "HEIF", "GIF", "APNG", "SVG"}
	// Raw 格式
	RawFormat = []string{"CR2", "NEF", "ARW", "ORF", "RAF", "DNG", "PEF", "SR2", "RW2", "X3F"}
)

// 获取指定文件夹下所有的文件夹
//
