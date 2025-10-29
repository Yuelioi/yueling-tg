package image

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"yueling_tg/internal/core/context"
	"yueling_tg/internal/message"
	"yueling_tg/pkg/config"
	"yueling_tg/pkg/plugin"

	"github.com/mymmrac/telego"
)

// 确保结构体实现接口
var _ plugin.Plugin = (*RandomGenerator)(nil)

// -------------------- 配置结构 --------------------

// PluginConfig 插件整体配置
type PluginConfig struct {
	DBPath       string           `mapstructure:"db_path"`
	ImagesFolder string           `mapstructure:"images_folder"`
	Categories   []CategoryConfig `mapstructure:"categories"`
}

// CategoryConfig 单个分类配置
type CategoryConfig struct {
	Commands        []string `mapstructure:"commands"`         // 触发命令列表，如 ["吃什么", "今天吃啥"]
	Folder          string   `mapstructure:"folder"`           // 对应文件夹
	Caption         string   `mapstructure:"caption"`          // 图片说明
	GridWidth       int      `mapstructure:"grid_width"`       // 宫格宽度
	Count           int      `mapstructure:"count"`            // 抽取数量（1=单图，≥4=宫格）
	MessageTemplate string   `mapstructure:"message_template"` // 消息模板
}

// -------------------- 图片哈希索引 --------------------

type ImageIndex struct {
	Hash     string `json:"hash"`
	Path     string `json:"path"`
	Category string `json:"category"`
	Filename string `json:"filename"`
}

type ImageIndexDB struct {
	Images map[string]*ImageIndex `json:"images"`
	mu     sync.RWMutex
}

// -------------------- 主结构体 --------------------

type RandomGenerator struct {
	*plugin.Base
	config  PluginConfig
	indexDB *ImageIndexDB
}

// -------------------- 插件入口 --------------------

func New() plugin.Plugin {
	info := &plugin.PluginInfo{
		ID:          "image",
		Name:        "图片管理",
		Description: "支持 吃什么 / 喝什么 / 玩什么 / 美少女 / 龙图 等指令",
		Version:     "1.3.0",
		Author:      "月离",
		Group:       "图库",
		Extra:       make(map[string]any),
	}

	rg := &RandomGenerator{
		indexDB: &ImageIndexDB{
			Images: make(map[string]*ImageIndex),
		},
		config: PluginConfig{},
	}

	// 获取配置
	defaultConfig := rg.getDefaultConfig()
	if err := config.GetPluginConfigOrDefault(info.ID, &rg.config, defaultConfig); err != nil {
		panic(fmt.Sprintf("加载插件配置失败: %v", err))
	}

	builder := plugin.New().Info(info)

	// 注册随机图片命令
	for _, cat := range rg.config.Categories {
		for _, cmd := range cat.Commands {
			builder.OnFullMatch(cmd).Do(func(c *context.Context) {
				rg.handleCommand(c, cat)
			})
		}
	}

	// 添加图片命令
	addCommands := rg.generateAddCommands()
	builder.OnCommand(addCommands...).Do(rg.handleAddImage)

	// 删除图片命令
	builder.OnCommand("删除图片").Block(true).Do(rg.handleDeleteImage)

	// 回调命令
	builder.OnCallbackStartsWith(info.ID).Priority(9).Do(rg.handleAnother)

	return builder.Go(rg)
}

func (rg *RandomGenerator) Init() error {
	if err := rg.loadOrCreateIndex(); err != nil {
		rg.Log.Error().Err(err).Msg("初始化图片索引失败")
	}
	return nil
}

// getDefaultConfig 获取默认配置
func (rg *RandomGenerator) getDefaultConfig() PluginConfig {

	return PluginConfig{
		DBPath: "./data/images/index.json",
		Categories: []CategoryConfig{
			{
				Commands:        []string{"吃什么", "今天吃啥"},
				Folder:          "吃的",
				Caption:         "今天吃这个吧！🍜",
				GridWidth:       750,
				Count:           4,
				MessageTemplate: "今天我们来点 %s 吧～ 😋",
			},
		},
	}
}

// generateAddCommands 根据配置生成添加图片命令
func (rg *RandomGenerator) generateAddCommands() []string {
	var commands []string
	for _, cat := range rg.config.Categories {
		commands = append(commands, "添加"+cat.Folder)
	}
	return commands
}

// findCategoryByFolder 根据文件夹名查找分类配置
func (rg *RandomGenerator) findCategoryByFolder(folder string) *CategoryConfig {
	for _, cat := range rg.config.Categories {
		if cat.Folder == folder {
			return &cat
		}
	}
	return nil
}

// -------------------- 再来一张 --------------------

func (rg *RandomGenerator) handleAnother(cmd string, c *context.Context) error {
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

	// 重新创建按钮
	buttons := rg.createButton(folder)

	params := &telego.EditMessageMediaParams{
		ChatID:      c.GetChatID(),
		MessageID:   msg.GetMessageID(),
		Media:       message.NewResource(imgPaths[0]).ToInputMedia(),
		ReplyMarkup: &buttons,
	}

	_, err = c.Api.EditMessageMedia(c.Ctx, params)
	if err != nil {
		rg.Log.Error().Err(err).Msg("编辑消息失败")
		c.AnswerCallback("换图失败 😢")
		return err
	}

	c.AnswerCallback("已换一张 🔄")
	return nil
}

func (rg *RandomGenerator) createButton(folder string) telego.InlineKeyboardMarkup {
	return telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			{
				telego.InlineKeyboardButton{
					Text:         "换一张 🔄",
					CallbackData: rg.PluginInfo().ID + "_" + folder,
				},
			},
		},
	}
}

// -------------------- 逻辑核心 --------------------

func (rg *RandomGenerator) handleCommand(c *context.Context, cfg CategoryConfig) {
	rg.Log.Debug().
		Str("from", c.GetUsername()).
		Strs("commands", cfg.Commands).
		Str("folder", cfg.Folder).
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

func (rg *RandomGenerator) sendSinglePhoto(c *context.Context, cfg CategoryConfig, imgPath string) {
	rg.Log.Debug().Str("file", imgPath).Msg("选取单图发送")

	photo := message.NewResource(imgPath).WithCaption(cfg.Caption)
	buttons := rg.createButton(cfg.Folder)

	c.SendPhotoWithMarkup(photo, buttons)
}

func (rg *RandomGenerator) sendMediaGroup(c *context.Context, cfg CategoryConfig, imgPaths []string) {
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
