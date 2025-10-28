package context

import (
	"fmt"
	"yueling_tg/internal/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ============ 基础消息发送方法 ============

// Send 发送文本消息
func (c *Context) Send(text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	return c.Api.Send(msg)
}

// Sendf 格式化发送文本消息
func (c *Context) Sendf(format string, args ...interface{}) (tgbotapi.Message, error) {
	return c.Send(fmt.Sprintf(format, args...))
}

// SendMarkdown 发送 Markdown 格式消息
func (c *Context) SendMarkdown(text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	msg.ParseMode = "Markdown"
	return c.Api.Send(msg)
}

// SendMarkdownV2 发送 MarkdownV2 格式消息
func (c *Context) SendMarkdownV2(text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	msg.ParseMode = "MarkdownV2"
	return c.Api.Send(msg)
}

func (c *Context) SendMessageWithMarkup(text string, markup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	msg.ReplyMarkup = &markup
	c.Api.Send(msg)
}

func (c *Context) SendPhotoWithMarkup(photo message.Resource, markup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewPhoto(c.GetChatID(), photo.Data)
	msg.Caption = photo.Caption
	msg.ReplyMarkup = &markup
	c.Api.Send(msg)
}

// SendHTML 发送 HTML 格式消息
func (c *Context) SendHTML(text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	msg.ParseMode = "HTML"
	return c.Api.Send(msg)
}

// SendWithKeyboard 发送带键盘的消息
func (c *Context) SendWithKeyboard(text string, keyboard interface{}) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	msg.ReplyMarkup = keyboard
	return c.Api.Send(msg)
}

// SendWithOptions 发送消息（完全自定义）
func (c *Context) SendWithOptions(text string, options func(*tgbotapi.MessageConfig)) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	if options != nil {
		options(&msg)
	}
	return c.Api.Send(msg)
}

// ============ 回复消息方法（Reply to Message）============

// Reply 回复当前消息
func (c *Context) Reply(text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// Replyf 格式化回复当前消息
func (c *Context) Replyf(format string, args ...interface{}) (tgbotapi.Message, error) {
	return c.Reply(fmt.Sprintf(format, args...))
}

// ReplyMarkdown 使用 Markdown 格式回复
func (c *Context) ReplyMarkdown(text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	msg.ReplyToMessageID = c.GetMessageID()
	msg.ParseMode = "Markdown"
	return c.Api.Send(msg)
}

// ReplyMarkdownV2 使用 MarkdownV2 格式回复
func (c *Context) ReplyMarkdownV2(text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	msg.ReplyToMessageID = c.GetMessageID()
	msg.ParseMode = "MarkdownV2"
	return c.Api.Send(msg)
}

// ReplyHTML 使用 HTML 格式回复
func (c *Context) ReplyHTML(text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	msg.ReplyToMessageID = c.GetMessageID()
	msg.ParseMode = "HTML"
	return c.Api.Send(msg)
}

// ReplyWithKeyboard 回复消息并带键盘
func (c *Context) ReplyWithKeyboard(text string, keyboard interface{}) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(c.GetChatID(), text)
	msg.ReplyToMessageID = c.GetMessageID()
	msg.ReplyMarkup = keyboard
	return c.Api.Send(msg)
}

// ============ 媒体发送方法 ============

// SendPhoto 发送图片
func (c *Context) SendPhoto(photo message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewPhoto(c.GetChatID(), photo.Data)
	msg.Caption = photo.Caption
	return c.Api.Send(msg)
}

// SendPhotoWithOptions 发送图片（自定义选项）
func (c *Context) SendPhotoWithOptions(photo message.Resource, options func(*tgbotapi.PhotoConfig)) (tgbotapi.Message, error) {
	msg := tgbotapi.NewPhoto(c.GetChatID(), photo.Data)
	if options != nil {
		options(&msg)
	}
	return c.Api.Send(msg)
}

// ReplyPhoto 回复图片
func (c *Context) ReplyPhoto(photo message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewPhoto(c.GetChatID(), photo.Data)
	msg.Caption = photo.Caption
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// SendVideo 发送视频
func (c *Context) SendVideo(video message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewVideo(c.GetChatID(), video.Data)
	msg.Caption = video.Caption
	return c.Api.Send(msg)
}

// SendVideoWithOptions 发送视频（自定义选项）
func (c *Context) SendVideoWithOptions(video message.Resource, options func(*tgbotapi.VideoConfig)) (tgbotapi.Message, error) {
	msg := tgbotapi.NewVideo(c.GetChatID(), video.Data)
	if options != nil {
		options(&msg)
	}
	return c.Api.Send(msg)
}

// ReplyVideo 回复视频
func (c *Context) ReplyVideo(video message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewVideo(c.GetChatID(), video.Data)
	msg.Caption = video.Caption
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// SendAnimation 发送动画/GIF
func (c *Context) SendAnimation(animation message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewAnimation(c.GetChatID(), animation.Data)
	msg.Caption = animation.Caption
	return c.Api.Send(msg)
}

// ReplyAnimation 回复动画/GIF
func (c *Context) ReplyAnimation(animation message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewAnimation(c.GetChatID(), animation.Data)
	msg.Caption = animation.Caption
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// SendDocument 发送文档
func (c *Context) SendDocument(document message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewDocument(c.GetChatID(), document.Data)
	msg.Caption = document.Caption
	return c.Api.Send(msg)
}

// SendDocumentWithOptions 发送文档（自定义选项）
func (c *Context) SendDocumentWithOptions(document message.Resource, options func(*tgbotapi.DocumentConfig)) (tgbotapi.Message, error) {
	msg := tgbotapi.NewDocument(c.GetChatID(), document.Data)
	if options != nil {
		options(&msg)
	}
	return c.Api.Send(msg)
}

// ReplyDocument 回复文档
func (c *Context) ReplyDocument(document message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewDocument(c.GetChatID(), document.Data)
	msg.Caption = document.Caption
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// SendAudio 发送音频
func (c *Context) SendAudio(audio message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewAudio(c.GetChatID(), audio.Data)
	msg.Caption = audio.Caption
	return c.Api.Send(msg)
}

// SendAudioWithOptions 发送音频（自定义选项）
func (c *Context) SendAudioWithOptions(audio message.Resource, options func(*tgbotapi.AudioConfig)) (tgbotapi.Message, error) {
	msg := tgbotapi.NewAudio(c.GetChatID(), audio.Data)
	if options != nil {
		options(&msg)
	}
	return c.Api.Send(msg)
}

// ReplyAudio 回复音频
func (c *Context) ReplyAudio(audio message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewAudio(c.GetChatID(), audio.Data)
	msg.Caption = audio.Caption
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// SendVoice 发送语音消息
func (c *Context) SendVoice(voice message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewVoice(c.GetChatID(), voice.Data)
	msg.Caption = voice.Caption
	return c.Api.Send(msg)
}

// ReplyVoice 回复语音消息
func (c *Context) ReplyVoice(voice message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewVoice(c.GetChatID(), voice.Data)
	msg.Caption = voice.Caption
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// SendVideoNote 发送圆形视频
func (c *Context) SendVideoNote(videoNote message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewVideoNote(c.GetChatID(), 0, videoNote.Data)
	return c.Api.Send(msg)
}

// ReplyVideoNote 回复圆形视频
func (c *Context) ReplyVideoNote(videoNote message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewVideoNote(c.GetChatID(), 0, videoNote.Data)
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// SendSticker 发送贴纸
func (c *Context) SendSticker(sticker message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewSticker(c.GetChatID(), sticker.Data)
	return c.Api.Send(msg)
}

// ReplySticker 回复贴纸
func (c *Context) ReplySticker(sticker message.Resource) (tgbotapi.Message, error) {
	msg := tgbotapi.NewSticker(c.GetChatID(), sticker.Data)
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// ============ 媒体组发送方法 ============

// SendMediaGroup 发送媒体组
func (c *Context) SendMediaGroup(resources []message.Resource, caption string) ([]tgbotapi.Message, error) {
	if len(resources) == 0 {
		return nil, fmt.Errorf("资源列表为空")
	}

	// Telegram Media Group 限制最多 10 张
	if len(resources) > 10 {
		resources = resources[:10]
	}

	var mediaGroup []interface{}

	for i, res := range resources {
		media := tgbotapi.NewInputMediaPhoto(res.Data)

		// Telegram 限制：只有第一张图片可以有 caption
		if i == 0 && caption != "" {
			media.Caption = caption
		}

		mediaGroup = append(mediaGroup, media)
	}

	msg := tgbotapi.NewMediaGroup(c.GetChatID(), mediaGroup)

	return c.Api.SendMediaGroup(msg)
}

// ReplyMediaGroup 回复媒体组
func (c *Context) ReplyMediaGroup(mediaGroup []interface{}, caption string) ([]tgbotapi.Message, error) {
	msg := tgbotapi.NewMediaGroup(c.GetChatID(), mediaGroup)
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.SendMediaGroup(msg)
}

func (c *Context) SendMediaGroupFromPaths(paths []string, caption string) ([]tgbotapi.Message, error) {
	resources := make([]message.Resource, len(paths))

	for i, path := range paths {
		resources[i] = message.NewResource(path)
		// 只给第一张图片添加 caption
		if i == 0 && caption != "" {
			resources[i].Caption = caption
		}
	}

	return c.SendMediaGroup(resources, caption)
}

// ============ 位置和联系人发送方法 ============

// SendLocation 发送位置
func (c *Context) SendLocation(latitude, longitude float64) (tgbotapi.Message, error) {
	msg := tgbotapi.NewLocation(c.GetChatID(), latitude, longitude)
	return c.Api.Send(msg)
}

// ReplyLocation 回复位置
func (c *Context) ReplyLocation(latitude, longitude float64) (tgbotapi.Message, error) {
	msg := tgbotapi.NewLocation(c.GetChatID(), latitude, longitude)
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// SendVenue 发送地点
func (c *Context) SendVenue(latitude, longitude float64, title, address string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewVenue(c.GetChatID(), title, address, latitude, longitude)
	return c.Api.Send(msg)
}

// ReplyVenue 回复地点
func (c *Context) ReplyVenue(latitude, longitude float64, title, address string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewVenue(c.GetChatID(), title, address, latitude, longitude)
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// SendContact 发送联系人
func (c *Context) SendContact(phoneNumber, firstName string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewContact(c.GetChatID(), phoneNumber, firstName)
	return c.Api.Send(msg)
}

// ReplyContact 回复联系人
func (c *Context) ReplyContact(phoneNumber, firstName string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewContact(c.GetChatID(), phoneNumber, firstName)
	msg.ReplyToMessageID = c.GetMessageID()
	return c.Api.Send(msg)
}

// SendPoll 发送投票
func (c *Context) SendPoll(question string, options []string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewPoll(c.GetChatID(), question, options...)
	return c.Api.Send(msg)
}

// SendQuiz 发送测验（带正确答案）
func (c *Context) SendQuiz(question string, options []string, correctOptionID int) (tgbotapi.Message, error) {
	msg := tgbotapi.NewPoll(c.GetChatID(), question, options...)
	msg.Type = "quiz"
	msg.CorrectOptionID = int64(correctOptionID)
	return c.Api.Send(msg)
}

// SendDice 发送骰子
func (c *Context) SendDice(emoji string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewDice(c.GetChatID())
	if emoji != "" {
		msg.Emoji = emoji
	}
	return c.Api.Send(msg)
}

// ============ 特殊操作方法 ============

// SendChatAction 发送聊天动作（如 typing, upload_photo 等）
func (c *Context) SendChatAction(action string) error {
	chatAction := tgbotapi.NewChatAction(c.GetChatID(), action)
	_, err := c.Api.Send(chatAction)
	return err
}

// SendTyping 发送正在输入动作
func (c *Context) SendTyping() error {
	return c.SendChatAction("typing")
}

// SendUploadPhoto 发送正在上传图片动作
func (c *Context) SendUploadPhoto() error {
	return c.SendChatAction("upload_photo")
}

// SendUploadVideo 发送正在上传视频动作
func (c *Context) SendUploadVideo() error {
	return c.SendChatAction("upload_video")
}

// SendUploadDocument 发送正在上传文档动作
func (c *Context) SendUploadDocument() error {
	return c.SendChatAction("upload_document")
}

// SendRecordVoice 发送正在录音动作
func (c *Context) SendRecordVoice() error {
	return c.SendChatAction("record_voice")
}

// SendRecordVideoNote 发送正在录制视频动作
func (c *Context) SendRecordVideoNote() error {
	return c.SendChatAction("record_video_note")
}

// ============ 消息编辑方法 ============

// EditMessageText 编辑消息文本
func (c *Context) EditMessageText(messageID int, text string) (tgbotapi.Message, error) {
	edit := tgbotapi.NewEditMessageText(c.GetChatID(), messageID, text)
	return c.Api.Send(edit)
}

// EditMessageTextWithOptions 编辑消息文本（自定义选项）
func (c *Context) EditMessageTextWithOptions(messageID int, text string, options func(*tgbotapi.EditMessageTextConfig)) (tgbotapi.Message, error) {
	edit := tgbotapi.NewEditMessageText(c.GetChatID(), messageID, text)
	if options != nil {
		options(&edit)
	}
	return c.Api.Send(edit)
}

// EditCurrentMessage 编辑当前消息（用于 CallbackQuery）
func (c *Context) EditCurrentMessage(text string) (tgbotapi.Message, error) {
	if c.Update.CallbackQuery == nil || c.Update.CallbackQuery.Message == nil {
		return tgbotapi.Message{}, fmt.Errorf("no callback query message to edit")
	}
	return c.EditMessageText(c.Update.CallbackQuery.Message.MessageID, text)
}

// EditMessageCaption 编辑媒体说明文本
func (c *Context) EditMessageCaption(messageID int, caption string) (tgbotapi.Message, error) {
	edit := tgbotapi.NewEditMessageCaption(c.GetChatID(), messageID, caption)
	return c.Api.Send(edit)
}

// EditMessageReplyMarkup 编辑消息的键盘
func (c *Context) EditMessageReplyMarkup(messageID int, keyboard *tgbotapi.InlineKeyboardMarkup) (tgbotapi.Message, error) {
	edit := tgbotapi.NewEditMessageReplyMarkup(c.GetChatID(), messageID, *keyboard)
	return c.Api.Send(edit)
}

// EditCurrentMessageWithKeyboard 编辑当前消息并更新键盘
func (c *Context) EditCurrentMessageWithKeyboard(text string, keyboard tgbotapi.InlineKeyboardMarkup) (tgbotapi.Message, error) {
	if c.Update.CallbackQuery == nil || c.Update.CallbackQuery.Message == nil {
		return tgbotapi.Message{}, fmt.Errorf("no callback query message to edit")
	}
	edit := tgbotapi.NewEditMessageText(
		c.GetChatID(),
		c.Update.CallbackQuery.Message.MessageID,
		text,
	)
	edit.ReplyMarkup = &keyboard
	return c.Api.Send(edit)
}

// ============ 消息删除方法 ============

// DeleteMessage 删除消息
func (c *Context) DeleteMessage(messageID int) error {
	delete := tgbotapi.NewDeleteMessage(c.GetChatID(), messageID)
	_, err := c.Api.Request(delete)
	return err
}

// DeleteCurrentMessage 删除当前消息
func (c *Context) DeleteCurrentMessage() error {
	return c.DeleteMessage(c.GetMessageID())
}

// ============ 回调查询响应方法 ============

// AnswerCallback 回答回调查询
func (c *Context) AnswerCallback(text string) error {
	if c.Update.CallbackQuery == nil {
		return fmt.Errorf("no callback query to answer")
	}
	callback := tgbotapi.NewCallback(c.Update.CallbackQuery.ID, text)
	_, err := c.Api.Request(callback)
	return err
}

// AnswerCallbackWithAlert 回答回调查询并显示警告弹窗
func (c *Context) AnswerCallbackWithAlert(text string) error {
	if c.Update.CallbackQuery == nil {
		return fmt.Errorf("no callback query to answer")
	}
	callback := tgbotapi.NewCallback(c.Update.CallbackQuery.ID, text)
	callback.ShowAlert = true
	_, err := c.Api.Request(callback)
	return err
}

// AnswerCallbackWithURL 回答回调查询并打开 URL
func (c *Context) AnswerCallbackWithURL(url string) error {
	if c.Update.CallbackQuery == nil {
		return fmt.Errorf("no callback query to answer")
	}
	callback := tgbotapi.NewCallback(c.Update.CallbackQuery.ID, "")
	callback.URL = url
	_, err := c.Api.Request(callback)
	return err
}

// ============ 内联查询响应方法 ============

// AnswerInlineQuery 回答内联查询
func (c *Context) AnswerInlineQuery(results []interface{}) error {
	if c.Update.InlineQuery == nil {
		return fmt.Errorf("no inline query to answer")
	}
	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: c.Update.InlineQuery.ID,
		Results:       results,
	}
	_, err := c.Api.Request(inlineConf)
	return err
}

// AnswerInlineQueryWithOptions 回答内联查询（自定义选项）
func (c *Context) AnswerInlineQueryWithOptions(results []interface{}, options func(*tgbotapi.InlineConfig)) error {
	if c.Update.InlineQuery == nil {
		return fmt.Errorf("no inline query to answer")
	}
	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: c.Update.InlineQuery.ID,
		Results:       results,
	}
	if options != nil {
		options(&inlineConf)
	}
	_, err := c.Api.Request(inlineConf)
	return err
}

// ============ 转发方法 ============

// ForwardMessage 转发消息
func (c *Context) ForwardMessage(toChatID int64, fromChatID int64, messageID int) (tgbotapi.Message, error) {
	forward := tgbotapi.NewForward(toChatID, fromChatID, messageID)
	return c.Api.Send(forward)
}

// ForwardCurrentMessage 转发当前消息到指定聊天
func (c *Context) ForwardCurrentMessage(toChatID int64) (tgbotapi.Message, error) {
	return c.ForwardMessage(toChatID, c.GetChatID(), c.GetMessageID())
}

// CopyMessage 复制消息（不显示转发来源）
func (c *Context) CopyMessage(toChatID int64, fromChatID int64, messageID int) (tgbotapi.MessageID, error) {
	copy := tgbotapi.NewCopyMessage(toChatID, fromChatID, messageID)
	return c.Api.CopyMessage(copy)
}

// CopyCurrentMessage 复制当前消息到指定聊天
func (c *Context) CopyCurrentMessage(toChatID int64) (tgbotapi.MessageID, error) {
	return c.CopyMessage(toChatID, c.GetChatID(), c.GetMessageID())
}

// ============ 群组管理方法 ============

// PinMessage 置顶消息
func (c *Context) PinMessage(messageID int, disableNotification bool) error {
	pin := tgbotapi.PinChatMessageConfig{
		ChatID:              c.GetChatID(),
		MessageID:           messageID,
		DisableNotification: disableNotification,
	}
	_, err := c.Api.Request(pin)
	return err
}

// PinCurrentMessage 置顶当前消息
func (c *Context) PinCurrentMessage(disableNotification bool) error {
	return c.PinMessage(c.GetMessageID(), disableNotification)
}

// UnpinMessage 取消置顶消息
func (c *Context) UnpinMessage(messageID int) error {
	unpin := tgbotapi.UnpinChatMessageConfig{
		ChatID:    c.GetChatID(),
		MessageID: messageID,
	}
	_, err := c.Api.Request(unpin)
	return err
}

// UnpinAllMessages 取消置顶所有消息
func (c *Context) UnpinAllMessages() error {
	unpin := tgbotapi.UnpinAllChatMessagesConfig{
		ChatID: c.GetChatID(),
	}
	_, err := c.Api.Request(unpin)
	return err
}

// BanChatMember 封禁群成员
func (c *Context) BanChatMember(userID int64) error {
	ban := tgbotapi.BanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: c.GetChatID(),
			UserID: userID,
		},
	}
	_, err := c.Api.Request(ban)
	return err
}

// UnbanChatMember 解封群成员
func (c *Context) UnbanChatMember(userID int64) error {
	unban := tgbotapi.UnbanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: c.GetChatID(),
			UserID: userID,
		},
	}
	_, err := c.Api.Request(unban)
	return err
}

// RestrictChatMember 限制群成员
func (c *Context) RestrictChatMember(userID int64, permissions tgbotapi.ChatPermissions) error {
	restrict := tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: c.GetChatID(),
			UserID: userID,
		},
		Permissions: &permissions,
	}
	_, err := c.Api.Request(restrict)
	return err
}

// PromoteChatMember 提升群成员为管理员
func (c *Context) PromoteChatMember(userID int64, privileges tgbotapi.ChatAdministratorsConfig) error {
	promote := tgbotapi.PromoteChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: c.GetChatID(),
			UserID: userID,
		},
	}
	_, err := c.Api.Request(promote)
	return err
}

// LeaveChat 离开群组/频道
func (c *Context) LeaveChat() error {
	leave := tgbotapi.LeaveChatConfig{
		ChatID: c.GetChatID(),
	}
	_, err := c.Api.Request(leave)
	return err
}
