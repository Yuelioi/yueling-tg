package message

import (
	"bytes"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Resource 封装发送资源（图片/文件等）的通用结构
type Resource struct {
	Caption string
	Data    tgbotapi.RequestFileData
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
			Data:    tgbotapi.FileURL(input),
			Caption: caption,
		}
	}

	// 2️⃣ 判断是否是本地文件（存在）
	if _, err := os.Stat(input); err == nil {
		return Resource{
			Data:    tgbotapi.FilePath(input),
			Caption: caption,
		}
	}

	// 3️⃣ 否则，默认为 Telegram file_id
	return Resource{
		Data:    tgbotapi.FileID(input),
		Caption: caption,
	}
}

// NewResourceFromBytes 用于直接从内存中构建
func NewResourceFromBytes(name string, data []byte) Resource {
	return NewResourceFromBytesWithCaption(name, data, "")
}

// NewResourceFromBytesWithCaption 从字节流构建带 caption 的资源
func NewResourceFromBytesWithCaption(name string, data []byte, caption string) Resource {
	reader := bytes.NewReader(data)
	return Resource{
		Data: tgbotapi.FileReader{
			Name:   name,
			Reader: reader,
		},
		Caption: caption,
	}
}

// WithCaption 链式设置 caption
func (r Resource) WithCaption(caption string) Resource {
	r.Caption = caption
	return r
}
