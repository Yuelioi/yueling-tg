package emotion

import (
	"io/fs"
	"math/rand"
	"path/filepath"
	"strings"

	"yueling_tg/internal/core/context"
	"yueling_tg/internal/message"
	"yueling_tg/pkg/plugin"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var _ plugin.Plugin = (*EmotePlugin)(nil)

type EmotePlugin struct {
	*plugin.Base
	path string
}

// -------------------- 插件入口 --------------------

func New() plugin.Plugin {
	ep := &EmotePlugin{
		Base: plugin.NewBase(&plugin.PluginInfo{
			ID:          "emote",
			Name:        "随机表情包",
			Description: "根据参数随机发送一张表情包/查询表情包列表",
			Version:     "1.0.0",
			Author:      "月离",
			Usage:       "#/##[关键词]",
			Group:       "图库",
			Extra:       make(map[string]any),
		}),
		path: "data/images/表情",
	}

	builder := plugin.New().
		Info(ep.PluginInfo())

	// 消息匹配器
	builder.OnMessage().
		Do(ep.emoteHandler)

	// 再来一张按钮（回调）
	builder.OnCallbackStartsWith(ep.PluginInfo().ID).
		Priority(9).
		Do(ep.another)

	return builder.Go()
}

// -------------------- 消息处理 --------------------

func (ep *EmotePlugin) emoteHandler(c *context.Context) {
	m := strings.TrimSpace(c.GetMessageText())
	if m == "" {
		return
	}

	switch {
	case m == "#": // 随机全部
		files, _ := ep.getEmoteFiles(nil)
		if len(files) == 0 {
			c.Reply("没有找到表情包 😢")
			return
		}
		ep.sendPhoto(c, files[rand.Intn(len(files))], "#")

	case strings.HasPrefix(m, "##"): // 查询列表
		query := strings.TrimSpace(strings.TrimPrefix(m, "##"))
		if query == "" {
			c.Reply("请输入查询关键词 😢")
			return
		}
		files, _ := ep.getEmoteFiles([]string{query})
		if len(files) == 0 {
			c.Reply("没有找到匹配的表情包 😢")
			return
		}

		names := make([]string, 0, len(files))
		for _, f := range files {
			names = append(names, strings.TrimSuffix(filepath.Base(f), filepath.Ext(f)))
		}
		limit := len(names)
		if limit > 20 {
			limit = 20
		}
		c.Reply("匹配到的表情包列表:\n" + strings.Join(names[:limit], "\n"))

	case strings.HasPrefix(m, "#"): // #关键词 → 随机匹配
		query := strings.TrimSpace(strings.TrimPrefix(m, "#"))
		var files []string
		if query == "" {
			files, _ = ep.getEmoteFiles(nil)
		} else {
			files, _ = ep.getEmoteFiles([]string{query})
		}
		if len(files) == 0 {
			c.Reply("没有找到匹配的表情包 😢")
			return
		}
		ep.sendPhoto(c, files[rand.Intn(len(files))], query)
	}
}

// -------------------- 再来一张 --------------------

func (ep *EmotePlugin) another(cmd string, c *context.Context) error {
	parts := strings.Split(cmd, "_")
	if len(parts) != 2 {
		ep.Log.Error().Str("cmd", cmd).Msg("按钮点击格式错误")
		return nil
	}
	query := parts[1]

	var files []string
	if query == "" || query == "#" {
		files, _ = ep.getEmoteFiles(nil)
	} else {
		files, _ = ep.getEmoteFiles([]string{query})
	}

	if len(files) == 0 {
		c.AnswerCallback("没有可用表情包 😢")
		return nil
	}

	choice := files[rand.Intn(len(files))]
	newPhoto := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(choice))

	// 获取原消息
	msg := c.GetCallbackQuery().Message
	if msg == nil {
		ep.Log.Error().Msg("callback没有原消息")
		return nil
	}

	// 创建按钮
	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("再来一张 🔄", ep.PluginInfo().ID+"_"+query),
		),
	)

	edit := tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      msg.Chat.ID,
			MessageID:   msg.MessageID,
			ReplyMarkup: &buttons,
		},
		Media: newPhoto,
	}

	c.Api.Send(edit)
	c.AnswerCallback("已换一张 🔄")
	return nil
}

// -------------------- 工具函数 --------------------

func (ep *EmotePlugin) sendPhoto(c *context.Context, file string, query string) {
	photo := message.NewResource(file)

	buttons := ep.createButton(query)

	c.SendPhotoWithMarkup(photo, buttons)
}

func (ep *EmotePlugin) createButton(query string) tgbotapi.InlineKeyboardMarkup {
	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("换一张 🔄", ep.PluginInfo().ID+"_"+query),
		),
	)

	return buttons
}

func (ep *EmotePlugin) getEmoteFiles(args []string) ([]string, error) {
	var files []string

	_ = filepath.WalkDir(ep.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
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

	return files, nil
}
