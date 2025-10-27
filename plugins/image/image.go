package image

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"yueling_tg/core/context"
	"yueling_tg/core/handler"
	"yueling_tg/core/message"
	"yueling_tg/core/on"
	"yueling_tg/core/plugin"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// 确保结构体实现接口
var _ plugin.Plugin = (*RandomGenerator)(nil)

// -------------------- 配置结构 --------------------

type RandomConfig struct {
	Command         string // 触发命令
	Folder          string // 对应文件夹
	Caption         string // 图片说明
	FontPath        string // 字体路径
	GridWidth       int    // 宫格宽度
	Count           int    // 抽取数量（1=单图，≥4=宫格）
	MessageTemplate string
}

// -------------------- 图片哈希索引 --------------------

type ImageIndex struct {
	Hash     string `json:"hash"`     // 文件 SHA1 哈希
	Path     string `json:"path"`     // 文件完整路径
	Category string `json:"category"` // 所属分类
	Filename string `json:"filename"` // 文件名
}

type ImageIndexDB struct {
	Images map[string]*ImageIndex `json:"images"` // key 是 hash
	mu     sync.RWMutex
}

// -------------------- 主结构体 --------------------

type RandomGenerator struct {
	*plugin.Base
	cfgs    []RandomConfig
	indexDB *ImageIndexDB
	dbPath  string
}

// -------------------- 插件入口 --------------------

func New() plugin.Plugin {
	font := "./data/fonts/华康新综艺简繁W7.ttf"

	rg := &RandomGenerator{
		Base: plugin.NewBase(&plugin.PluginInfo{
			ID:          "random",
			Name:        "随机图片生成器",
			Description: "支持 吃什么 / 喝什么 / 玩什么 / 美少女 / 龙图 等指令",
			Version:     "1.2.0",
			Author:      "月离",
			Group:       "funny",
			Extra:       make(map[string]any),
		}),
		dbPath: "./data/images/index.json",

		cfgs: []RandomConfig{
			{"吃什么", "吃的", "今天吃这个吧！🍜", font, 750, 4, "今天我们来点 %s 吧～ 😋"},
			{"喝什么", "喝的", "喝一杯？☕", font, 750, 4, "来杯 %s 吧～ ☕"},
			{"玩什么", "玩的", "玩这个吧～ 🎮", font, 750, 4, "玩玩 %s 怎么样～ 🎮"},
			{"来点零食", "零食", "来点小零食吧 🍪", font, 750, 4, "尝尝 %s 吧～ 🍪"},
			{"我老婆呢", "老婆", "", font, 750, 1, ""},
			{"我老公呢", "老公", "", font, 750, 1, ""},
			{"美少女", "美少女", "", font, 750, 1, ""},
			{"龙图", "龙图", "", font, 750, 1, ""},
			{"福瑞", "福瑞", "", font, 750, 1, ""},
			{"杂鱼", "杂鱼", "", font, 750, 1, ""},
			{"ba", "ba", "", font, 750, 1, ""},
		},
	}

	// 初始化索引数据库
	rg.indexDB = &ImageIndexDB{
		Images: make(map[string]*ImageIndex),
	}

	// 加载或创建索引
	if err := rg.loadOrCreateIndex(); err != nil {
		rg.Log.Error().Err(err).Msg("初始化图片索引失败")
	}

	matchers := []*plugin.Matcher{}

	// 原来的随机图片命令
	for _, cfg := range rg.cfgs {
		c := cfg
		m := on.OnFullMatch([]string{c.Command}, handler.NewHandler(func(ctx *context.Context) {
			rg.handleCommand(ctx, c)
		}))
		matchers = append(matchers, m)
	}

	// 添加图片命令
	addCommands := []string{
		"添加老婆", "添加老公", "添加龙图", "添加福瑞", "添加杂鱼",
		"添加吃的", "添加喝的", "添加玩的", "添加零食", "添加美少女", "添加ba",
	}
	addHandler := handler.NewHandler(rg.handleAddImage)
	m2 := on.OnCommand(addCommands, true, addHandler)
	matchers = append(matchers, m2)

	// 删除图片命令
	deleteCommands := []string{"删除图片"}
	deleteHandler := handler.NewHandler(rg.handleDeleteImage)
	m3 := on.OnCommand(deleteCommands, true, deleteHandler)
	matchers = append(matchers, m3)

	h4 := handler.NewHandler(rg.another)
	m4 := on.OnCallbackStartsWith([]string{rg.PluginInfo().ID}, h4).SetPriority(9)

	matchers = append(matchers, m4)

	rg.SetMatchers(matchers)

	return rg
}

func (rg *RandomGenerator) another(cmd string, c *context.Context) error {
	rg.Log.Debug().Str("from", cmd).Msg("收到随机按钮点击")

	parts := strings.Split(cmd, "_")
	if len(parts) != 2 {
		rg.Log.Error().Str("cmd", cmd).Msg("按钮点击格式错误")
		return nil
	}

	folder := parts[1]

	// 获取原消息 ID
	msg := c.GetCallbackQuery().Message
	if msg == nil {
		rg.Log.Error().Msg("callback没有原消息")
		return nil
	}

	// 重新选择一张图片
	imgPaths, err := rg.selectImages(filepath.Join("./data/images", folder), 1)
	if err != nil {
		c.AnswerCallback("没有可用图片 😢")
		return nil
	}

	newPhoto := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(imgPaths[0]))

	// 重新创建按钮
	buttons := rg.createButton(folder)

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

func (rg *RandomGenerator) createButton(folder string) tgbotapi.InlineKeyboardMarkup {
	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("换一张 🔄", rg.PluginInfo().ID+"_"+folder),
		),
	)

	return buttons
}

// -------------------- 逻辑核心 --------------------

func (rg *RandomGenerator) handleCommand(c *context.Context, cfg RandomConfig) {
	rg.Log.Debug().
		Str("from", c.GetUsername()).
		Str("cmd", cfg.Command).
		Msg("收到随机命令")

	folder := filepath.Join("./data/images", cfg.Folder)
	imgPaths, err := rg.selectImages(folder, cfg.Count)
	if err != nil {
		rg.Log.Error().Err(err).Str("folder", folder).Msg("无法读取图片")
		c.Reply("还没准备好图片哦～ 📂")
		return
	}

	if cfg.Count == 1 {
		rg.sendSinglePhoto(c, cfg, imgPaths[0])
		return
	}

	rg.sendMediaGroup(c, cfg, imgPaths)
}

func (rg *RandomGenerator) selectImages(folder string, count int) ([]string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	var imgPaths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp":
			imgPaths = append(imgPaths, filepath.Join(folder, e.Name()))
		}
	}

	if len(imgPaths) == 0 {
		return nil, fmt.Errorf("文件夹里没有图片")
	}

	// 随机打乱
	rand.Shuffle(len(imgPaths), func(i, j int) { imgPaths[i], imgPaths[j] = imgPaths[j], imgPaths[i] })

	if count > len(imgPaths) {
		count = len(imgPaths)
	}

	return imgPaths[:count], nil
}

func (rg *RandomGenerator) sendSinglePhoto(c *context.Context, cfg RandomConfig, imgPath string) {
	rg.Log.Debug().Str("file", imgPath).Msg("选取单图发送")

	photo := message.NewResource(imgPath).WithCaption(cfg.Caption)

	buttons := rg.createButton(cfg.Command)

	c.SendPhotoWithMarkup(photo, buttons)
}

func (rg *RandomGenerator) sendMediaGroup(c *context.Context, cfg RandomConfig, imgPaths []string) {
	n := len(imgPaths)
	resources := make([]message.Resource, n)
	var names []string

	for i, p := range imgPaths {
		filename := filepath.Base(p)
		pureName := strings.TrimSuffix(filename, filepath.Ext(filename))

		resources[i] = message.NewResource(p)
		names = append(names, pureName)
	}

	caption := cfg.Caption
	if cfg.MessageTemplate != "" {
		caption = fmt.Sprintf(cfg.MessageTemplate, strings.Join(names, ", "))
	}

	c.SendMediaGroup(resources, caption)
}
