package random

import (
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"yueling_tg/core/context"
	"yueling_tg/core/handler"
	"yueling_tg/core/message"
	"yueling_tg/core/on"
	"yueling_tg/core/plugin"
)

var _ plugin.Plugin = (*EmotePlugin)(nil)

func NewEmotePlugin() plugin.Plugin {
	ep := &EmotePlugin{
		Base: plugin.NewBase(&plugin.PluginInfo{
			ID:          "emote",
			Name:        "随机表情包",
			Description: "根据参数随机发送一张表情包/查询表情包列表",
			Version:     "1.0.0",
			Author:      "月离",
			Usage:       "#/##[关键词]",
			Group:       "funny",
			Extra:       make(map[string]any),
		}),
		path: "data/images/表情",
	}

	handler := handler.NewHandler(ep.emoteHandler)
	m := on.OnMessage(handler)
	ep.AddMatcher(m)

	return ep
}

type EmotePlugin struct {
	*plugin.Base
	path string
}

// ----------------------------------------------
// 获取表情包
// ----------------------------------------------
func (ep *EmotePlugin) emoteHandler(c *context.Context) {
	m := strings.TrimSpace(c.GetMessageText())
	if m == "" {
		return
	}

	switch {
	case m == "#": // 随机全部
		files, err := ep.getEmoteFiles(nil)
		if err != nil || len(files) == 0 {
			c.Reply("没有找到表情包 😢")
			return
		}
		choice := files[rand.Intn(len(files))]
		data, _ := os.ReadFile(choice)
		c.SendPhoto(message.NewResourceFromBytes(filepath.Base(choice), data))

	case strings.HasPrefix(m, "##"): // 查询列表
		query := strings.TrimPrefix(m, "##")
		query = strings.TrimSpace(query)
		if query == "" {
			c.Reply("请输入查询关键词 😢")
			return
		}

		files, err := ep.getEmoteFiles([]string{query})
		if err != nil || len(files) == 0 {
			c.Reply("没有找到匹配的表情包 😢")
			return
		}

		names := make([]string, 0, len(files))
		for _, f := range files {
			names = append(names, strings.TrimSuffix(filepath.Base(f), filepath.Ext(f)))
		}
		c.Reply("匹配到的表情包列表:\n" + strings.Join(names[:20], "\n"))

	case strings.HasPrefix(m, "#"): // #关键词 → 随机匹配
		query := strings.TrimPrefix(m, "#")
		query = strings.TrimSpace(query)

		var files []string
		var err error
		if query == "" {
			// 直接随机全部
			files, err = ep.getEmoteFiles(nil)
		} else {
			files, err = ep.getEmoteFiles([]string{query})
		}

		if err != nil || len(files) == 0 {
			c.Reply("没有找到匹配的表情包 😢")
			return
		}

		choice := files[rand.Intn(len(files))]
		data, _ := os.ReadFile(choice)
		if query == "" {
			c.SendPhoto(message.NewResourceFromBytes(filepath.Base(choice), data))
		} else {
			c.SendPhoto(message.NewResourceFromBytes(filepath.Base(choice), data))
		}

	default:
		// 不处理
		return
	}
}

// ----------------------------------------------
// 获取表情包列表
// ----------------------------------------------

func (ep *EmotePlugin) getEmoteFiles(args []string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(ep.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// 如果有参数就匹配文件名
		if len(args) > 0 {
			for _, arg := range args {
				if strings.Contains(strings.ToLower(d.Name()), strings.ToLower(arg)) {
					files = append(files, path)
					break
				}
			}
		} else {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}
