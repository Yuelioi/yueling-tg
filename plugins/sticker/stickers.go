package sticker

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strconv"
	"strings"
	"time"

	"yueling_tg/internal/core/context"
	"yueling_tg/internal/message"

	"github.com/chai2010/webp"
	"github.com/mymmrac/telego"
	"golang.org/x/image/draw"
)

// -------------------- 命令处理 --------------------

func (sp *StickerPlugin) handleAddSticker(c *context.Context) {
	userID := c.GetUserID()

	stickerSets := sp.db.Sets
	if len(stickerSets) == 0 {
		c.Reply("你还没有创建任何贴纸库，请先使用 '创建贴纸集' 命令创建")
		return
	}

	// 设置用户状态
	sp.stateMutex.Lock()
	sp.userStates[userID] = &UserState{
		Mode:       "waiting",
		UserID:     userID,
		LastUpdate: time.Now(),
	}
	sp.stateMutex.Unlock()

	// 显示贴纸库选择
	sp.showStickerSetSelection(c, stickerSets)
}

// -------------------- 显示贴纸库选择 --------------------

func (sp *StickerPlugin) showStickerSetSelection(c *context.Context, stickerSets []*StickerSetData) {
	userID := c.GetUserID()
	var buttons [][]telego.InlineKeyboardButton

	for _, set := range stickerSets {
		buttonText := fmt.Sprintf("📦 %s", set.Title)
		callbackData := fmt.Sprintf("%s:select:%s:%d", sp.PluginInfo().ID, set.Name, userID)

		buttons = append(buttons, []telego.InlineKeyboardButton{{
			Text:         buttonText,
			CallbackData: callbackData,
		}})
	}

	// 添加取消按钮
	buttons = append(buttons, []telego.InlineKeyboardButton{{
		Text:         "❌ 取消",
		CallbackData: fmt.Sprintf("%s:cancel:%d", sp.PluginInfo().ID, userID),
	}})

	markup := telego.InlineKeyboardMarkup{InlineKeyboard: buttons}
	c.SendMessageWithMarkup("请选择要添加贴纸的贴纸库：", markup)
}

// -------------------- 贴纸库选择处理 --------------------

func (sp *StickerPlugin) handleStickerSetSelect(cmd string, c *context.Context) error {
	// 格式: pluginID:select:SET_NAME:INITIATOR_ID
	parts := strings.Split(cmd, ":")
	if len(parts) < 4 {
		c.AnswerCallback("参数错误")
		return nil
	}

	stickerSetName := parts[2]
	initiatorID, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		c.AnswerCallback("参数错误")
		return nil
	}

	clickerID := c.GetUserID()

	// ✅ 仅允许发起者本人操作
	if state, ok := sp.userStates[initiatorID]; ok {
		if clickerID != state.UserID {
			c.AnswerCallback("只有发起者可以操作")
			return nil
		}
	}

	// 获取贴纸库信息
	params := &telego.GetStickerSetParams{Name: stickerSetName}
	stickerSet, err := c.Api.GetStickerSet(c.Ctx, params)
	if err != nil {
		c.AnswerCallback("获取贴纸库信息失败")
		return nil
	}

	// 更新状态
	sp.stateMutex.Lock()
	sp.userStates[initiatorID] = &UserState{
		Mode:               "adding",
		StickerSetID:       stickerSetName,
		StickerSetTitle:    stickerSet.Title,
		UserID:             initiatorID,
		LastUpdate:         time.Now(),
		ProcessingMsgID:    0,
		ProcessingChatID:   0,
		ProcessingInFlight: false,
	}
	sp.stateMutex.Unlock()

	// 编辑提示消息，加入取消按钮（允许发起者随时取消）
	msg := c.GetCallbackQuery().Message
	if msg != nil {
		cancelButton := telego.InlineKeyboardMarkup{
			InlineKeyboard: [][]telego.InlineKeyboardButton{{
				{
					Text:         "❌ 取消添加",
					CallbackData: fmt.Sprintf("%s:cancel:%d", sp.PluginInfo().ID, initiatorID),
				},
			}},
		}
		editParams := &telego.EditMessageTextParams{
			ChatID:      c.GetChatID(),
			MessageID:   msg.GetMessageID(),
			Text:        fmt.Sprintf("✅ 已选择贴纸库：%s\n\n现在请发送图片，我会自动添加到该贴纸库。若要放弃，请点击下方取消。", stickerSet.Title),
			ReplyMarkup: &cancelButton,
		}
		c.Api.EditMessageText(c.Ctx, editParams)
	}
	c.AnswerCallback("请发送图片")
	return nil
}

// -------------------- 取消处理 --------------------

func (sp *StickerPlugin) handleCancel(cmd string, c *context.Context) error {
	// 格式: pluginID:cancel:INITIATOR_ID
	parts := strings.Split(cmd, ":")
	if len(parts) < 3 {
		c.AnswerCallback("参数错误")
		return nil
	}

	initiatorID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		c.AnswerCallback("参数错误")
		return nil
	}

	clickerID := c.GetUserID()

	// ✅ 仅允许发起者本人取消
	sp.stateMutex.RLock()
	state, ok := sp.userStates[initiatorID]
	sp.stateMutex.RUnlock()
	if ok {
		if clickerID != state.UserID {
			c.AnswerCallback("只有发起者可以取消该操作")
			return nil
		}
	}

	// 如果存在正在使用的处理消息，尝试把那条消息编辑为“已取消”
	if ok && state.ProcessingMsgID != 0 {
		processingMsg := &telego.Message{
			MessageID: state.ProcessingMsgID,
			Chat:      telego.Chat{ID: state.ProcessingChatID},
		}
		sp.editMessage(c, processingMsg, "❌ 已取消添加贴纸")
	}

	// 删除状态
	sp.stateMutex.Lock()
	delete(sp.userStates, initiatorID)
	sp.stateMutex.Unlock()

	msg := c.GetCallbackQuery().Message
	if msg != nil {
		params := &telego.EditMessageTextParams{
			ChatID:    c.GetChatID(),
			MessageID: msg.GetMessageID(),
			Text:      "❌ 已取消添加贴纸",
		}
		c.Api.EditMessageText(c.Ctx, params)
	}

	c.AnswerCallback("已取消")
	return nil
}

// -------------------- 图片处理 --------------------

func (sp *StickerPlugin) handleImage(c *context.Context) {
	userID := c.GetUserID()

	// 获取状态
	sp.stateMutex.RLock()
	state, exists := sp.userStates[userID]
	sp.stateMutex.RUnlock()

	if !exists || state.Mode != "adding" {
		return
	}

	// 校验归属
	if state.UserID != userID {
		c.Reply("只有发起者可以上传贴纸图片")
		return
	}

	// 超时检查
	if time.Since(state.LastUpdate) > 10*time.Minute {
		// 如果有 processing 消息，编辑为超时提示
		sp.stateMutex.Lock()
		if state.ProcessingMsgID != 0 {
			processingMsg := &telego.Message{
				MessageID: state.ProcessingMsgID,
				Chat:      telego.Chat{ID: state.ProcessingChatID},
			}
			sp.editMessage(c, processingMsg, "⌛ 添加贴纸已超时，操作已取消")
		}
		delete(sp.userStates, userID)
		sp.stateMutex.Unlock()

		c.Reply("添加贴纸已超时，请重新开始")
		return
	}

	// 获取图片 fileID
	photo, ok := c.GetPhoto()
	if !ok {
		// 如果不是图片，忽略
		return
	}

	// 确保只有一个协程在操作该用户的 processing 消息（避免并发竞争）
	// 使用简单的 in-flight 标志
	sp.stateMutex.Lock()
	// re-get state pointer (to be safe in concurrent env)
	state, exists = sp.userStates[userID]
	if !exists || state.Mode != "adding" {
		sp.stateMutex.Unlock()
		return
	}

	// 如果没有 processing 消息则发送一条新的并记录
	if state.ProcessingMsgID == 0 {
		msg, _ := c.Reply("⏳ 正在处理图片 1 ...")
		if msg != nil {
			state.ProcessingMsgID = msg.MessageID
			state.ProcessingChatID = msg.Chat.ID
		}
		// 初始化计数
		state.ProcessCount = 1
	} else {
		// 已有处理消息，增加计数
		state.ProcessCount++
		// 构造用于 edit 的 message 对象
	}
	// 标记 in-flight
	state.ProcessingInFlight = true
	processingMsgID := state.ProcessingMsgID
	processingChatID := state.ProcessingChatID
	currentCount := state.ProcessCount
	// 更新时间戳
	state.LastUpdate = time.Now()
	sp.stateMutex.Unlock()

	// 构造 processing message 对象（用于 edit）
	var processingMsg *telego.Message
	if processingMsgID != 0 {
		processingMsg = &telego.Message{
			MessageID: processingMsgID,
			Chat:      telego.Chat{ID: processingChatID},
		}
	}

	// 先把状态消息更新为正在处理（原地替换）
	if processingMsg != nil {
		sp.editMessage(c, processingMsg, fmt.Sprintf("⏳ 正在处理图片 %d ...", currentCount))
	}

	// 执行耗时处理（转换并添加贴纸）
	webpData, err := sp.downloadAndConvertToWebP(c, photo)
	if err != nil {
		// 编辑状态消息为失败并清除 in-flight
		if processingMsg != nil {
			sp.editMessage(c, processingMsg, fmt.Sprintf("❌ 处理图片失败：%v", err))
		}
		sp.stateMutex.Lock()
		if s, ok := sp.userStates[userID]; ok {
			s.ProcessingInFlight = false
		}
		sp.stateMutex.Unlock()
		return
	}

	// 调用添加 API
	if err := sp.addStickerToSet(c, userID, state.StickerSetID, webpData); err != nil {
		if processingMsg != nil {
			sp.editMessage(c, processingMsg, fmt.Sprintf("❌ 添加贴纸失败：%v", err))
		}
		sp.stateMutex.Lock()
		if s, ok := sp.userStates[userID]; ok {
			s.ProcessingInFlight = false
		}
		sp.stateMutex.Unlock()
		return
	}

	// 添加成功：更新状态消息为已添加第 X 张
	if processingMsg != nil {
		sp.editMessage(c, processingMsg, fmt.Sprintf("✅ 已添加第 %d 张贴纸\n贴纸库：%s\n\n继续发送图片或点击取消", currentCount, state.StickerSetTitle))
	}

	// 清除 in-flight 标志并更新时间
	sp.stateMutex.Lock()
	if s, ok := sp.userStates[userID]; ok {
		s.ProcessingInFlight = false
		s.LastUpdate = time.Now()
	}
	sp.stateMutex.Unlock()
}

// -------------------- 辅助函数 --------------------

func (sp *StickerPlugin) downloadAndConvertToWebP(c *context.Context, fileID string) ([]byte, error) {
	params := &telego.GetFileParams{FileID: fileID}
	file, err := c.Api.GetFile(c.Ctx, params)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	fileURL := c.Api.FileDownloadURL(file.FilePath)
	resp, err := sp.httpClient.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("下载图片失败: %v", err)
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取图片失败: %v", err)
	}

	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("解码图片失败: %v", err)
	}
	sp.Log.Debug().Msgf("原始图片格式: %s, 尺寸: %dx%d", format, img.Bounds().Dx(), img.Bounds().Dy())

	resizedImg := sp.resizeImageForSticker(img)

	var buf bytes.Buffer
	err = webp.Encode(&buf, resizedImg, &webp.Options{Lossless: true, Quality: 90})
	if err != nil {
		return nil, fmt.Errorf("WebP编码失败: %v", err)
	}
	sp.Log.Debug().Msgf("WebP转换完成, 大小: %d bytes", buf.Len())

	return buf.Bytes(), nil
}

func (sp *StickerPlugin) resizeImageForSticker(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	maxSize := 512

	var newWidth, newHeight int
	if width > height {
		newWidth = maxSize
		newHeight = height * maxSize / width
	} else {
		newHeight = maxSize
		newWidth = width * maxSize / height
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

func (sp *StickerPlugin) addStickerToSet(c *context.Context, userID int64, stickerSetName string, webpData []byte) error {
	inputSticker := telego.InputSticker{
		Sticker:   telego.InputFile{File: message.NewNameReader("sticker.webp", webpData)},
		EmojiList: []string{"😀"},
		Format:    "static",
	}

	params := &telego.AddStickerToSetParams{
		UserID:  userID,
		Name:    stickerSetName,
		Sticker: inputSticker,
	}

	if err := c.Api.AddStickerToSet(c.Ctx, params); err != nil {
		sp.Log.Error().Err(err).Msg("添加贴纸失败")
		return err
	}
	return nil
}

func (sp *StickerPlugin) editMessage(c *context.Context, msg *telego.Message, text string) {
	if msg == nil {
		return
	}
	params := &telego.EditMessageTextParams{
		ChatID:    telego.ChatID{ID: msg.Chat.ID},
		MessageID: msg.MessageID,
		Text:      text,
	}
	c.Api.EditMessageText(c.Ctx, params)
}
