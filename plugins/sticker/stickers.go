package sticker

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"

	"github.com/chai2010/webp"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/image/draw"
)

var _ plugin.Plugin = (*StickerPlugin)(nil)

// -------------------- 数据结构 --------------------

// 用户状态
type UserState struct {
	Mode            string // "waiting" - 等待选择贴纸库, "adding" - 正在添加贴纸
	StickerSetID    string // 选中的贴纸库ID (short_name)
	StickerSetTitle string // 贴纸库标题
	LastUpdate      time.Time
}

type UserStickerSet struct {
	UserID    int64
	ShortName string // 贴纸库的短名称 (例如: "setname")
	Title     string // 贴纸库的标题
	IsCreator bool   // 是否为创建者
}

// -------------------- 插件主结构 --------------------

type StickerPlugin struct {
	*plugin.Base
	userStates map[int64]*UserState // userID -> state (用于临时的添加流程状态)
	stateMutex sync.RWMutex
	httpClient *http.Client

	db     *StickerSetDB
	dbPath string
}

func New() plugin.Plugin {
	sp := &StickerPlugin{
		Base: plugin.NewBase(&plugin.PluginInfo{
			ID:          "sticker",
			Name:        "贴纸管理",
			Description: "管理Telegram贴纸库",
			Version:     "1.0.0",
			Author:      "月离",
			Usage:       "添加贴纸",
			Group:       "工具",
		}),
		userStates: make(map[int64]*UserState),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		db:         &StickerSetDB{},
		dbPath:     "./data/botsticker.json",
	}

	// 加载数据
	if err := sp.loadData(); err != nil {
		sp.Log.Warn().Err(err).Msg("加载贴纸集数据失败，使用空数据库")
	} else {
		sp.Log.Info().Msgf("加载了 %d 个贴纸集", len(sp.db.Sets))
	}

	builder := plugin.New().
		Info(sp.PluginInfo())

	builder.OnCommand("创建贴纸集").Block(true).Do(sp.handleCreateSet)
	builder.OnCommand("删除贴纸集").Do(sp.handleDeleteSet)
	builder.OnCommand("查看贴纸集").Do(sp.handleListSet)

	// 添加贴纸命令
	builder.OnStartsWith("添加贴纸").
		Do(sp.handleAddSticker)

	// 处理贴纸库选择
	builder.OnCallbackStartsWith(sp.PluginInfo().ID + ":select:").
		Priority(9).
		Do(sp.handleStickerSetSelect)

	// 处理取消
	builder.OnCallbackStartsWith(sp.PluginInfo().ID + ":cancel").
		Priority(9).
		Do(sp.handleCancel)

	// 处理图片消息
	builder.OnMessage().
		Priority(8).
		Do(sp.handleImage)

	return builder.Go()
}

// -------------------- 命令处理 --------------------

func (sp *StickerPlugin) handleAddSticker(c *context.Context) {
	userID := c.GetUserID()

	// 获取用户的所有贴纸库
	stickerSets := sp.db.Sets

	if len(stickerSets) == 0 {
		c.Reply("你还没有创建任何贴纸库，请先在Telegram中创建贴纸库")
		return
	}

	// 设置用户状态
	sp.stateMutex.Lock()
	sp.userStates[userID] = &UserState{
		Mode:       "waiting",
		LastUpdate: time.Now(),
	}
	sp.stateMutex.Unlock()

	// 创建贴纸库选择按钮
	sp.showStickerSetSelection(c, stickerSets)
}

// -------------------- 显示贴纸库选择 --------------------

func (sp *StickerPlugin) showStickerSetSelection(c *context.Context, stickerSets []*StickerSetData) {
	var buttons [][]tgbotapi.InlineKeyboardButton

	for _, set := range stickerSets {
		buttonText := fmt.Sprintf("📦 %s", set.Title)
		callbackData := fmt.Sprintf("%s:select:%s", sp.PluginInfo().ID, set.Name)

		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData),
		))
	}

	// 添加取消按钮
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("❌ 取消", sp.PluginInfo().ID+":cancel"),
	))

	markup := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	c.SendMessageWithMarkup("请选择要添加贴纸的贴纸库：", markup)
}

// -------------------- 贴纸库选择处理 --------------------

func (sp *StickerPlugin) handleStickerSetSelect(cmd string, c *context.Context) error {
	userID := c.GetUserID()

	// 格式: sticker:select:SET_NAME
	parts := strings.Split(cmd, ":")
	if len(parts) < 3 {
		c.AnswerCallback("参数错误")
		return nil
	}

	stickerSetName := parts[2]

	// 获取贴纸库信息
	getStickerSet := tgbotapi.GetStickerSetConfig{
		Name: stickerSetName,
	}

	stickerSet, err := c.Api.GetStickerSet(getStickerSet)
	if err != nil {
		c.AnswerCallback("获取贴纸库信息失败")
		return nil
	}

	// 更新用户状态
	sp.stateMutex.Lock()
	sp.userStates[userID] = &UserState{
		Mode:            "adding",
		StickerSetID:    stickerSetName,
		StickerSetTitle: stickerSet.Title,
		LastUpdate:      time.Now(),
	}
	sp.stateMutex.Unlock()

	// 编辑消息
	msg := c.GetCallbackQuery().Message
	if msg != nil {
		// 创建取消按钮
		cancelButton := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("❌ 取消添加", sp.PluginInfo().ID+":cancel"),
			),
		)

		edit := tgbotapi.EditMessageTextConfig{
			BaseEdit: tgbotapi.BaseEdit{
				ChatID:      msg.Chat.ID,
				MessageID:   msg.MessageID,
				ReplyMarkup: &cancelButton,
			},
			Text: fmt.Sprintf("✅ 已选择贴纸库：%s\n\n现在请发送图片，我会自动添加到该贴纸库", stickerSet.Title),
		}
		c.Api.Send(edit)
	}

	c.AnswerCallback("请发送图片")
	return nil
}

// -------------------- 取消处理 --------------------

func (sp *StickerPlugin) handleCancel(cmd string, c *context.Context) error {
	userID := c.GetUserID()

	// 清除用户状态
	sp.stateMutex.Lock()
	delete(sp.userStates, userID)
	sp.stateMutex.Unlock()

	msg := c.GetCallbackQuery().Message
	if msg != nil {
		edit := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, "❌ 已取消添加贴纸")
		c.Api.Send(edit)
	}

	c.AnswerCallback("已取消")
	return nil
}

// -------------------- 图片处理 --------------------

func (sp *StickerPlugin) handleImage(c *context.Context) {
	userID := c.GetUserID()

	// 检查用户状态
	sp.stateMutex.RLock()
	state, exists := sp.userStates[userID]
	sp.stateMutex.RUnlock()

	if !exists || state.Mode != "adding" {
		return // 不处理
	}

	// 检查是否超时（10分钟）
	if time.Since(state.LastUpdate) > 10*time.Minute {
		sp.stateMutex.Lock()
		delete(sp.userStates, userID)
		sp.stateMutex.Unlock()
		c.Reply("添加贴纸已超时，请重新开始")
		return
	}

	// 下载并转换图片
	statusMsg, _ := c.Reply("正在处理图片...")

	photo, ok := c.GetPhoto()
	if !ok {
		sp.editMessage(c, &statusMsg, "❌ 请发送一张图片")
		return
	}
	webpData, err := sp.downloadAndConvertToWebP(c, photo)
	if err != nil {
		sp.editMessage(c, &statusMsg, fmt.Sprintf("❌ 处理图片失败：%v", err))
		return
	}

	// 添加贴纸到贴纸库
	err = sp.addStickerToSet(c, userID, state.StickerSetID, webpData)
	if err != nil {
		sp.editMessage(c, &statusMsg, fmt.Sprintf("❌ 添加贴纸失败：%v", err))
		return
	}

	// 更新状态时间
	sp.stateMutex.Lock()
	if s, ok := sp.userStates[userID]; ok {
		s.LastUpdate = time.Now()
	}
	sp.stateMutex.Unlock()

	// 显示成功消息和取消按钮
	cancelButton := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ 取消添加", sp.PluginInfo().ID+":cancel"),
		),
	)

	edit := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      statusMsg.Chat.ID,
			MessageID:   statusMsg.MessageID,
			ReplyMarkup: &cancelButton,
		},
		Text: fmt.Sprintf("✅ 添加成功！\n贴纸库：%s\n\n继续发送图片或点击取消", state.StickerSetTitle),
	}
	c.Api.Send(edit)

}

// -------------------- 辅助函数 --------------------

// 下载并转换图片为 WebP
func (sp *StickerPlugin) downloadAndConvertToWebP(c *context.Context, fileID string) ([]byte, error) {
	// 获取文件URL
	fileConfig := tgbotapi.FileConfig{FileID: fileID}
	file, err := c.Api.GetFile(fileConfig)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	fileURL := file.Link(c.Api.Token)

	// 下载图片
	resp, err := sp.httpClient.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("下载图片失败: %v", err)
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取图片失败: %v", err)
	}

	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %v", err)
	}

	sp.Log.Debug().Msgf("原始图片格式: %s, 尺寸: %dx%d", format, img.Bounds().Dx(), img.Bounds().Dy())

	// 调整图片尺寸（贴纸要求512px的一边）
	resizedImg := sp.resizeImageForSticker(img)

	// 转换为 WebP
	var buf bytes.Buffer
	err = webp.Encode(&buf, resizedImg, &webp.Options{
		Lossless: true,
		Quality:  90,
	})
	if err != nil {
		return nil, fmt.Errorf("WebP编码失败: %v", err)
	}

	sp.Log.Debug().Msgf("WebP转换完成, 大小: %d bytes", buf.Len())

	return buf.Bytes(), nil
}

// 调整图片尺寸以符合贴纸要求
func (sp *StickerPlugin) resizeImageForSticker(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 贴纸要求：最长边为512px
	maxSize := 512
	var newWidth, newHeight int

	if width > height {
		newWidth = maxSize
		newHeight = height * maxSize / width
	} else {
		newHeight = maxSize
		newWidth = width * maxSize / height
	}

	// 创建新图片
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	return dst
}

// 添加贴纸到贴纸库
func (sp *StickerPlugin) addStickerToSet(c *context.Context, userID int64, stickerSetName string, webpData []byte) error {
	fileBytes := tgbotapi.FileBytes{
		Name:  "sticker.webp",
		Bytes: webpData,
	}

	addSticker := tgbotapi.AddStickerConfig{
		UserID:     userID,
		Name:       stickerSetName,
		PNGSticker: fileBytes,
		Emojis:     "😀",
	}

	// ===== 调试打印 =====
	fmt.Println("=== AddSticker 调试信息 ===")
	fmt.Printf("UserID: %d\n", addSticker.UserID)
	fmt.Printf("StickerSet Name: %s\n", addSticker.Name)
	fmt.Printf("Sticker Size: %d bytes\n", len(webpData))
	fmt.Println("============================")

	_, err := c.Api.Request(addSticker)
	if err != nil {
		fmt.Printf("添加贴纸失败: %v\n", err)
	}

	return err
}

// 编辑消息
func (sp *StickerPlugin) editMessage(c *context.Context, msg *tgbotapi.Message, text string) {
	if msg == nil {
		return
	}
	edit := tgbotapi.NewEditMessageText(msg.Chat.ID, msg.MessageID, text)
	c.Api.Send(edit)
}
