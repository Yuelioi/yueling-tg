package sticker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"yueling_tg/internal/core/context"
	"yueling_tg/internal/message"
	"yueling_tg/pkg/plugin/params"

	"github.com/mymmrac/telego"
)

// -------------------- 数据结构 --------------------

type StickerSetData struct {
	Name  string `json:"name"`  // 短名 pack_xxx_by_bot
	Title string `json:"title"` // 显示名称
}

type StickerSetDB struct {
	Sets []*StickerSetData `json:"sets"`
}

// -------------------- 命令处理 --------------------

func (sp *StickerPlugin) handleCreateSet(c *context.Context, cmdCtx params.CommandContext) {
	if cmdCtx.Args.Len() < 2 {
		c.Reply("❌ 用法：创建贴纸集 <显示名称> <短名称>")
		return
	}
	title := cmdCtx.Args.Get(0)
	shortName := cmdCtx.Args.Get(1) // 用户指定的短名称

	bot, err := c.GetBot()
	if err != nil {
		c.Replyf("❌ 获取机器人信息失败: %v", err)
		return
	}

	setName := fmt.Sprintf("pack_%s_by_%s", shortName, bot.Username)

	// 检查是否已存在

	checkParams := &telego.GetStickerSetParams{
		Name: setName,
	}
	_, err = c.Api.GetStickerSet(c.Ctx, checkParams)
	if err == nil {
		c.Replyf("❌ 贴纸集 '%s' 已经存在了", setName)
		return
	}

	photo, ok := c.GetPhoto()
	if !ok {
		c.Reply("❌ 请发送一张图片作为第一张贴纸")
		return
	}

	// 转为合法图片
	webpData, err := sp.downloadAndConvertToWebP(c, photo)
	if err != nil {
		c.Replyf("❌ 转换图片失败: %v", err)
		return
	}

	// 创建贴纸集
	ok, err = sp.createStickerSet(c, c.GetUserID(), setName, title, webpData)
	if err != nil || !ok {
		c.Replyf("❌ 创建贴纸集失败: %v", err)
		return
	}

	// 保存记录到 db
	sp.db.Sets = append(sp.db.Sets, &StickerSetData{
		Name:  setName,
		Title: title,
	})
	sp.saveData()

	c.Replyf("✅ 成功创建贴纸集：%s\n\n🔗 链接：https://t.me/addstickers/%s", title, setName)
}

func (sp *StickerPlugin) createStickerSet(c *context.Context, userID int64, setName, setTitle string, webpData []byte) (bool, error) {
	// 创建 InputSticker
	inputSticker := telego.InputSticker{
		Sticker: telego.InputFile{
			File: message.NewNameReader("sticker.webp", webpData),
		},
		Format:    "static", // 静态贴纸
		EmojiList: []string{"😀"},
	}

	params := &telego.CreateNewStickerSetParams{
		UserID:      userID,
		Name:        setName,
		Title:       setTitle,
		Stickers:    []telego.InputSticker{inputSticker},
		StickerType: "regular", // 静态贴纸
	}

	// 创建贴纸集
	err := c.Api.CreateNewStickerSet(c.Ctx, params)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (sp *StickerPlugin) handleDeleteSet(ctx *context.Context, cmdCtx params.CommandContext) {
	if cmdCtx.Args.Len() < 1 {
		ctx.Reply("❌ 用法：删除贴纸集 <短名>")
		return
	}

	name := cmdCtx.Args.Get(0)

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

func (sp *StickerPlugin) handleListSet(ctx *context.Context) {
	if len(sp.db.Sets) == 0 {
		ctx.Reply("当前没有任何贴纸集")
		return
	}

	var sb strings.Builder
	sb.WriteString("📝 当前 Bot 管理的贴纸集：\n\n")

	for i, s := range sp.db.Sets {
		link := fmt.Sprintf("https://t.me/addstickers/%s", s.Name)
		sb.WriteString(fmt.Sprintf("%d. %s\n🔗 [%s](%s)\n\n", i+1, s.Title, s.Name, link))
	}

	ctx.ReplyMarkdown(sb.String())
}

// -------------------- 数据管理 --------------------

func (sp *StickerPlugin) loadData() error {
	data, err := os.ReadFile(sp.config.DBPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, sp.db)
}

func (sp *StickerPlugin) saveData() error {
	dir := filepath.Dir(sp.config.DBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(sp.db, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(sp.config.DBPath, data, 0644)
}
