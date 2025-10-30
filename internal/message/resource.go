package message

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mymmrac/telego"
)

// Resource 封装发送资源（图片/文件等）的通用结构
type Resource struct {
	Caption string
	Data    telego.InputFile
}

// NewResource 自动识别类型并构建 Resource
// 输入可以是：
// - 本地路径: "images/pic.jpg"
// - 网络URL: "https://example.com/pic.jpg"
// - Telegram file_id: "AgACAgUAAx..." 等
func NewResource(input string) Resource {
	return NewResourceWithCaption(input, "")
}

// NewResourceWithCaption 带 caption 的资源构造器
func NewResourceWithCaption(input string, caption string) Resource {
	// 1️⃣ 判断是否是 URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return Resource{
			Data:    telego.InputFile{FileID: input},
			Caption: caption,
		}
	}

	// 2️⃣ 判断是否是本地文件（存在）
	if info, err := os.Stat(input); err == nil && !info.IsDir() {
		data, err := os.ReadFile(input)
		if err == nil {
			return Resource{
				Data: telego.InputFile{
					File: NewNameReader(filepath.Base(input), data),
				},
				Caption: caption,
			}
		}
	}

	// 3️⃣ 否则，默认为 Telegram file_id
	return Resource{
		Data:    telego.InputFile{FileID: input},
		Caption: caption,
	}
}

func (r Resource) WithCaption(caption string) Resource {
	r.Caption = caption
	return r
}
func (r Resource) ToInputFile() telego.InputFile {
	return r.Data
}

func (r Resource) ToInputMedia() telego.InputMedia {
	ext := strings.ToLower(filepath.Ext(r.getName()))

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return &telego.InputMediaPhoto{
			Type:    "photo",
			Media:   r.Data,
			Caption: r.Caption,
		}
	case ".mp4", ".mov", ".mkv":
		return &telego.InputMediaVideo{
			Type:    "video",
			Media:   r.Data,
			Caption: r.Caption,
		}
	case ".mp3", ".wav", ".ogg":
		return &telego.InputMediaAudio{
			Type:    "audio",
			Media:   r.Data,
			Caption: r.Caption,
		}
	default:
		return &telego.InputMediaDocument{
			Type:    "document",
			Media:   r.Data,
			Caption: r.Caption,
		}
	}
}
func (r Resource) getName() string {
	// 优先取 File.Name()（如果存在）
	if r.Data.File != nil {
		if n, ok := r.Data.File.(interface{ Name() string }); ok {
			return n.Name()
		}
	}
	// 否则用 FileID（可能是路径或 URL）
	return r.Data.FileID
}

// NewResourceFromBytes 用于直接从内存中构建
func NewResourceFromBytes(name string, data []byte) Resource {
	return NewResourceFromBytesWithCaption(name, data, "")
}

// NewResourceFromBytesWithCaption 从字节流构建带 caption 的资源
type NamedReader struct {
	name   string
	reader io.Reader
}

func NewNameReader(name string, data []byte) *NamedReader {
	reader := bytes.NewReader(data)

	return &NamedReader{
		name:   name,
		reader: reader,
	}
}

func (n *NamedReader) Read(p []byte) (int, error) {
	return n.reader.Read(p)
}

func (n *NamedReader) Name() string {
	return n.name
}

// NewResourceFromBytesWithCaption 从字节流构建带 caption 的资源
func NewResourceFromBytesWithCaption(name string, data []byte, caption string) Resource {
	return Resource{
		Data: telego.InputFile{
			File: NewNameReader(name, data),
		},
		Caption: caption,
	}
}
