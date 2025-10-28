package sticker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/params"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var _ plugin.Plugin = (*BotStickerPlugin)(nil)

// -------------------- 数据结构 --------------------

type StickerSetData struct {
	Name  string `json:"name"`  // 短名 pack_xxx_by_bot
	Title string `json:"title"` // 显示名称
}

type StickerSetDB struct {
	Sets []*StickerSetData `json:"sets"`
	mu   sync.RWMutex      `json:"-"`
}

// -------------------- 插件结构 --------------------

type BotStickerPlugin struct {
	*plugin.Base
}

// -------------------- 命令处理 --------------------

func (sp *StickerPlugin) handleCreateSet(c *context.Context, cmdCtx params.CommandContext) {
	if cmdCtx.Args.Len() < 2 {
		c.Reply("❌ 用法：创建贴纸集 <显示名称> <短名称> ")
		return
	}
	title := cmdCtx.Args.Get(0)
	shortName := cmdCtx.Args.Get(1) // 用户指定的短名称

	bot, err := c.Api.GetMe()
	if err != nil {
		c.Reply("❌ 获取BotID失败")
		return
	}

	setName := fmt.Sprintf("pack_%s_by_%s", shortName, bot.UserName)

	// 检查是否已存在
	sp.db.mu.Lock()
	_, err = c.Api.GetStickerSet(tgbotapi.GetStickerSetConfig{
		Name: setName,
	})
	if err == nil {
		c.Replyf("贴纸集 '%s' 已经存在了", setName)
	}
	// 如果错误不是因为贴纸集不存在，则返回错误
	if tgErr, ok := err.(tgbotapi.Error); ok && tgErr.Code != 400 {
		c.Replyf("贴纸集检查失败: %v", err)
	}

	photo, ok := c.GetPhoto()
	if !ok {
		c.Reply("请发送一张图片")
		return
	}
	// 转为合法图片
	webpData, err := sp.downloadAndConvertToWebP(c, photo)
	if err != nil {
		c.Replyf("转换图片失败: %v", err)
		return
	}

	sp.createStickerSet(c.Api, c.GetUserID(), setName, title, webpData)

	// 保存记录到 db
	sp.db.Sets = append(sp.db.Sets, &StickerSetData{
		Name:  setName,
		Title: title,
	})
	sp.db.mu.Unlock()
	sp.saveData()

	c.Replyf("✅ 将创建贴纸集：%s (%s)，请发送第一张贴纸以完成创建", title, setName)
}

func (sp *StickerPlugin) createStickerSet(bot *tgbotapi.BotAPI, userID int64, setName, setTitle string, webpData []byte) (bool, error) {

	// 准备贴纸文件
	fileBytes := tgbotapi.FileBytes{
		Name:  "sticker.webp",
		Bytes: webpData,
	}

	stickerConfig := tgbotapi.NewStickerSetConfig{
		UserID:     userID,
		Name:       setName,
		Title:      setTitle,
		PNGSticker: fileBytes,
		Emojis:     "😀",
	}

	// 创建贴纸集
	resp, err := bot.Request(stickerConfig)
	if err != nil {
		return false, err
	}

	if resp.Ok {
		return true, nil
	}
	return false, fmt.Errorf("创建贴纸集失败: %s", resp.Description)
}

func (sp *StickerPlugin) handleDeleteSet(ctx *context.Context, cmdCtx params.CommandContext) {
	if cmdCtx.Args.Len() < 1 {
		ctx.Reply("❌ 用法：删除贴纸集 <短名>")
		return
	}

	name := cmdCtx.Args.Get(0)
	sp.db.mu.Lock()
	defer sp.db.mu.Unlock()

	index := -1
	for i, s := range sp.db.Sets {
		if s.Name == name {
			index = i
			break
		}
	}

	if index == -1 {
		ctx.Reply("❌ 未找到该贴纸集")
		return
	}

	sp.db.Sets = append(sp.db.Sets[:index], sp.db.Sets[index+1:]...)
	sp.saveData()
	ctx.Reply("✅ 删除成功")
}

func (sp *StickerPlugin) handleListSet(ctx *context.Context, cmdCtx params.CommandContext) {
	sp.db.mu.RLock()
	defer sp.db.mu.RUnlock()

	if len(sp.db.Sets) == 0 {
		ctx.Reply("当前没有任何贴纸集")
		return
	}

	msg := "📝 当前Bot管理的贴纸集：\n"
	for _, s := range sp.db.Sets {
		msg += fmt.Sprintf("%s (%s)\n", s.Title, s.Name)
	}

	ctx.Reply(msg)
}

// -------------------- 数据管理 --------------------

func (sp *StickerPlugin) loadData() error {
	data, err := os.ReadFile(sp.dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, sp.db)
}

func (sp *StickerPlugin) saveData() error {
	dir := filepath.Dir(sp.dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	sp.db.mu.RLock()
	data, err := json.MarshalIndent(sp.db, "", "  ")
	sp.db.mu.RUnlock()
	if err != nil {
		return err
	}

	return os.WriteFile(sp.dbPath, data, 0644)
}
