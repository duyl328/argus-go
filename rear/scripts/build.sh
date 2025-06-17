#!/bin/bash

# 构建和打包脚本
set -e

PROJECT_NAME="image-processor"
VERSION="1.0.0"
BUILD_DIR="build"
TOOLS_DIR="tools"

# 清理构建目录
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# 构建不同平台的可执行文件
platforms=("windows/amd64" "linux/amd64" "darwin/amd64" "darwin/arm64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}

    output_name=$PROJECT_NAME
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    output_dir="$BUILD_DIR/${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    mkdir -p $output_dir

    echo "Building for $GOOS/$GOARCH..."

    # 构建 Go 程序
    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_dir/$output_name .

    # 复制工具文件
    bin_dir="$output_dir/bin"
    mkdir -p $bin_dir

    if [ $GOOS = "windows" ]; then
        # Windows 工具
        cp $TOOLS_DIR/windows/magick.exe $bin_dir/
        cp $TOOLS_DIR/windows/exiftool.exe $bin_dir/
        # 复制可能需要的 DLL 文件
        cp $TOOLS_DIR/windows/*.dll $bin_dir/ 2>/dev/null || true
    elif [ $GOOS = "linux" ]; then
        # Linux 工具
        cp $TOOLS_DIR/linux/magick $bin_dir/
        cp $TOOLS_DIR/linux/exiftool $bin_dir/
        chmod +x $bin_dir/*
    elif [ $GOOS = "darwin" ]; then
        # macOS 工具
        if [ $GOARCH = "arm64" ]; then
            cp $TOOLS_DIR/darwin-arm64/magick $bin_dir/
            cp $TOOLS_DIR/darwin-arm64/exiftool $bin_dir/
        else
            cp $TOOLS_DIR/darwin-amd64/magick $bin_dir/
            cp $TOOLS_DIR/darwin-amd64/exiftool $bin_dir/
        fi
        chmod +x $bin_dir/*
    fi

    # 创建压缩包
    cd $BUILD_DIR
    if [ $GOOS = "windows" ]; then
        zip -r "${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}.zip" "${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    else
        tar -czf "${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}.tar.gz" "${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    fi
    cd ..

    echo "Built $output_dir"
done

echo "All builds completed!"

# 下载工具的脚本部分（单独运行）
download_tools() {
    echo "Downloading external tools..."
    mkdir -p $TOOLS_DIR/{windows,linux,darwin-amd64,darwin-arm64}

    # 下载 ImageMagick (示例 URL，需要根据实际情况调整)
    # Windows
    wget -O imagemagick-windows.zip "https://imagemagick.org/download/binaries/ImageMagick-7.1.0-portable-Q16-x64.zip"
    unzip imagemagick-windows.zip -d $TOOLS_DIR/windows/

    # 下载 ExifTool
    # Windows
    wget -O exiftool-windows.zip "https://exiftool.org/exiftool-12.50.zip"
    unzip exiftool-windows.zip -d temp/
    cp temp/exiftool(-k).exe $TOOLS_DIR/windows/exiftool.exe

    echo "Tools downloaded!"
}

# 如果需要下载工具，取消注释下面这行
# download_tools