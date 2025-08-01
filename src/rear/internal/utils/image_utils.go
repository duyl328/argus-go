package utils

/*
统一图像格式
输入图像路径，判断图像是否是 PNG、JPG、Webp
如果是 Raw 格式，首先转换为 PNG 格式作为中转格式（首个版本可不支持）
如果是复杂格式则使用 ImageMagick 转换为 PNG 中转格式，再转换为 JPG 或 Webp 格式
如果是 PNG、JPG、Webp 则将其通过 LibVips 转换为 JPG 或 Webp 格式
*/

/*
压缩图片（支持输入格式为 jpg.webp） 输出为 jpg.webp
输入内容为指定图片的原图路径
*/
