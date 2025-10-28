package message

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/fogleman/gg"
)

// DrawGrid 绘制宫格图片
// files: 图片路径列表
// rows, cols: 行列数
// gridWidth: 整个宫格宽度
// fontPath: 字体路径（ttf文件）
func DrawGrid(files []string, rows, cols, gridWidth int, fontPath string) ([]byte, error) {
	squareSize := gridWidth / cols
	gridHeight := squareSize * rows

	dc := gg.NewContext(gridWidth, gridHeight)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	if fontPath != "" {
		if err := dc.LoadFontFace(fontPath, 36); err != nil {
			return nil, err
		}
	}

	for idx, path := range files {
		if idx >= rows*cols {
			break
		}

		imgFile, err := os.Open(path)
		if err != nil {
			continue
		}
		defer imgFile.Close()

		srcImg, _, err := image.Decode(imgFile)
		if err != nil {
			continue
		}

		bounds := srcImg.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		minSide := width
		if height < width {
			minSide = height
		}

		left := (width - minSide) / 2
		top := (height - minSide) / 2
		cropRect := image.Rect(left, top, left+minSide, top+minSide)
		cropped := cropImage(srcImg, cropRect)

		resized := resizeImage(cropped, squareSize, squareSize)

		x := (idx % cols) * squareSize
		y := (idx / cols) * squareSize

		dc.DrawImage(resized, x, y)

		if fontPath != "" {
			name := filepath.Base(path)
			name = strings.TrimSuffix(name, filepath.Ext(name))
			dc.SetRGB(1, 1, 1)
			dc.DrawStringAnchored(name, float64(x+40), float64(y+40), 0, 0)
			dc.SetRGB(0.3, 0.3, 0.3)
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// cropImage 从原图裁剪矩形区域
func cropImage(img image.Image, rect image.Rectangle) image.Image {
	dst := image.NewRGBA(rect.Sub(rect.Min))
	draw.Draw(dst, dst.Bounds(), img, rect.Min, draw.Src)
	return dst
}

// resizeImage 使用 gg 缩放图像
func resizeImage(img image.Image, width, height int) image.Image {
	dc := gg.NewContext(width, height)
	dc.DrawImageAnchored(img, width/2, height/2, 0.5, 0.5)
	return dc.Image()
}
