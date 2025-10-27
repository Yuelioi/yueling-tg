package jm

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/chai2010/webp"
)

// calculateBlockNum 计算图片的block_num
func calculateBlockNum(scrambleId, id int64, filename string) uint32 {
	if id < scrambleId {
		return 0
	} else if id < 268850 {
		return 10
	}

	var x uint32
	if id < 421926 {
		x = 10
	} else {
		x = 8
	}

	s := fmt.Sprintf("%d%s", id, filename)
	s = md5Hex(s)

	// 获取最后一个字符的ASCII值
	lastChar := s[len(s)-1]
	blockNum := uint32(lastChar) % x
	blockNum = blockNum*2 + 2

	return blockNum
}

// detectImageFormat 检测图片格式
func detectImageFormat(data []byte) (ImageFormat, error) {
	if len(data) < 12 {
		return "", fmt.Errorf("图片数据太短")
	}

	// 检测 GIF (47 49 46 38)
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return ImageFormatGIF, nil
	}

	// 检测 PNG (89 50 4E 47)
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return ImageFormatPNG, nil
	}

	// 检测 JPEG (FF D8 FF)
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return ImageFormatJPEG, nil
	}

	// 检测 WebP (52 49 46 46 ... 57 45 42 50)
	if len(data) >= 12 &&
		data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return ImageFormatWEBP, nil
	}

	return "", fmt.Errorf("未知的图片格式")
}

// decodeImage 解码图片
func decodeImage(data []byte, format ImageFormat) (image.Image, error) {
	reader := strings.NewReader(string(data))

	switch format {
	case ImageFormatJPEG:
		return jpeg.Decode(reader)
	case ImageFormatPNG:
		return png.Decode(reader)
	case ImageFormatWEBP:
		return webp.Decode(reader)
	case ImageFormatGIF:
		return gif.Decode(reader)
	default:
		return nil, fmt.Errorf("不支持的图片格式: %s", format)
	}
}

// saveImage 保存图片
func saveImage(img image.Image, savePath string, format ImageFormat) error {
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	switch format {
	case ImageFormatJPEG:
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
	case ImageFormatPNG:
		return png.Encode(file, img)
	case ImageFormatWEBP:
		return webp.Encode(file, img, &webp.Options{Lossless: false, Quality: 95})
	case ImageFormatGIF:
		return gif.Encode(file, img, nil)
	default:
		return fmt.Errorf("不支持的输出格式: %s", format)
	}
}

// stitchImage 拼接被分割的图片
func stitchImage(srcImg image.Image, blockNum uint32) image.Image {
	bounds := srcImg.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 创建新图片
	stitchedImg := image.NewRGBA(bounds)

	// 计算余数高度
	remainderHeight := height % int(blockNum)
	blockHeight := height / int(blockNum)

	for i := uint32(0); i < blockNum; i++ {
		// 计算当前块的高度
		currentBlockHeight := blockHeight
		srcYStart := height - (blockHeight * int(i+1)) - remainderHeight
		dstYStart := blockHeight * int(i)

		// 第一块需要加上余数高度
		if i == 0 {
			currentBlockHeight += remainderHeight
		} else {
			dstYStart += remainderHeight
		}

		// 逐行复制像素
		for y := 0; y < currentBlockHeight; y++ {
			srcY := srcYStart + y
			dstY := dstYStart + y

			for x := 0; x < width; x++ {
				stitchedImg.Set(x, dstY, srcImg.At(x, srcY))
			}
		}
	}

	return stitchedImg
}

// Helper functions
func md5Hex(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
