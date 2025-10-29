package context

import (
	"fmt"
	"yueling_tg/internal/message"

	"github.com/mymmrac/telego"
)

// ============ 基础消息发送方法 ============

// Send 发送文本消息
func (c *Context) Send(text string) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID: c.GetChatID(),
		Text:   text,
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// Sendf 格式化发送文本消息
func (c *Context) Sendf(format string, args ...interface{}) (*telego.Message, error) {
	return c.Send(fmt.Sprintf(format, args...))
}

// SendMarkdown 发送 Markdown 格式消息
func (c *Context) SendMarkdown(text string) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID:    c.GetChatID(),
		Text:      text,
		ParseMode: telego.ModeMarkdown,
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// SendMarkdownV2 发送 MarkdownV2 格式消息
func (c *Context) SendMarkdownV2(text string) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID:    c.GetChatID(),
		Text:      text,
		ParseMode: telego.ModeMarkdownV2,
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// SendHTML 发送 HTML 格式消息
func (c *Context) SendHTML(text string) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID:    c.GetChatID(),
		Text:      text,
		ParseMode: telego.ModeHTML,
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// SendWithKeyboard 发送带键盘的消息（ReplyKeyboardMarkup 或 InlineKeyboardMarkup）
func (c *Context) SendWithKeyboard(text string, keyboard telego.ReplyMarkup) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID:      c.GetChatID(),
		Text:        text,
		ReplyMarkup: keyboard,
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// SendMessageWithMarkup 发送带内联键盘的消息
func (c *Context) SendMessageWithMarkup(text string, markup telego.InlineKeyboardMarkup) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID:      c.GetChatID(),
		Text:        text,
		ReplyMarkup: &markup,
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// SendPhotoWithMarkup 发送带内联键盘的图片消息
func (c *Context) SendPhotoWithMarkup(photo message.Resource, markup telego.InlineKeyboardMarkup) (*telego.Message, error) {
	msg := telego.SendPhotoParams{
		ChatID:      c.GetChatID(),
		Photo:       photo.Data,
		Caption:     photo.Caption,
		ReplyMarkup: &markup,
	}
	return c.Api.SendPhoto(c.Ctx, &msg)
}

// SendWithOptions 发送消息（完全自定义）
func (c *Context) SendWithOptions(text string, options func(msg *telego.SendMessageParams)) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID: c.GetChatID(),
		Text:   text,
	}
	if options != nil {
		options(&msg)
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// ============ 媒体发送方法 ============

// SendPhoto 发送图片
func (c *Context) SendPhoto(photo message.Resource) (*telego.Message, error) {
	msg := telego.SendPhotoParams{
		ChatID:  c.GetChatID(),
		Photo:   photo.Data,
		Caption: photo.Caption,
	}
	return c.Api.SendPhoto(c.Ctx, &msg)
}

// SendPhotoWithOptions 发送图片（自定义选项）
func (c *Context) SendPhotoWithOptions(photo message.Resource, options func(*telego.SendPhotoParams)) (*telego.Message, error) {
	msg := telego.SendPhotoParams{
		ChatID: c.GetChatID(),
		Photo:  photo.Data,
	}
	if options != nil {
		options(&msg)
	}
	return c.Api.SendPhoto(c.Ctx, &msg)
}

// ReplyPhoto 回复图片
func (c *Context) ReplyPhoto(photo message.Resource) (*telego.Message, error) {
	msg := telego.SendPhotoParams{
		ChatID:  c.GetChatID(),
		Photo:   photo.Data,
		Caption: photo.Caption,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
	}
	return c.Api.SendPhoto(c.Ctx, &msg)
}

// SendVideo 发送视频
func (c *Context) SendVideo(video message.Resource) (*telego.Message, error) {
	msg := telego.SendVideoParams{
		ChatID:  c.GetChatID(),
		Video:   video.Data,
		Caption: video.Caption,
	}
	return c.Api.SendVideo(c.Ctx, &msg)
}

// SendVideoWithOptions 发送视频（自定义选项）
func (c *Context) SendVideoWithOptions(video message.Resource, options func(*telego.SendVideoParams)) (*telego.Message, error) {
	msg := telego.SendVideoParams{
		ChatID: c.GetChatID(),
		Video:  video.Data,
	}
	if options != nil {
		options(&msg)
	}
	return c.Api.SendVideo(c.Ctx, &msg)
}

// SendAnimation 发送动画/GIF
func (c *Context) SendAnimation(animation message.Resource) (*telego.Message, error) {
	msg := telego.SendAnimationParams{
		ChatID:    c.GetChatID(),
		Animation: animation.Data,
		Caption:   animation.Caption,
	}
	return c.Api.SendAnimation(c.Ctx, &msg)
}

// SendDocument 发送文档
func (c *Context) SendDocument(document message.Resource) (*telego.Message, error) {
	msg := telego.SendDocumentParams{
		ChatID:   c.GetChatID(),
		Document: document.Data,
		Caption:  document.Caption,
	}
	return c.Api.SendDocument(c.Ctx, &msg)
}

// SendDocumentWithOptions 发送文档（自定义选项）
func (c *Context) SendDocumentWithOptions(document message.Resource, options func(*telego.SendDocumentParams)) (*telego.Message, error) {
	msg := telego.SendDocumentParams{
		ChatID:   c.GetChatID(),
		Document: document.Data,
	}
	if options != nil {
		options(&msg)
	}
	return c.Api.SendDocument(c.Ctx, &msg)
}

// SendAudio 发送音频
func (c *Context) SendAudio(audio message.Resource) (*telego.Message, error) {
	msg := telego.SendAudioParams{
		ChatID:  c.GetChatID(),
		Audio:   audio.Data,
		Caption: audio.Caption,
	}
	return c.Api.SendAudio(c.Ctx, &msg)
}

// SendAudioWithOptions 发送音频（自定义选项）
func (c *Context) SendAudioWithOptions(audio message.Resource, options func(*telego.SendAudioParams)) (*telego.Message, error) {
	msg := telego.SendAudioParams{
		ChatID: c.GetChatID(),
		Audio:  audio.Data,
	}
	if options != nil {
		options(&msg)
	}
	return c.Api.SendAudio(c.Ctx, &msg)
}

// SendVoice 发送语音消息
func (c *Context) SendVoice(voice message.Resource) (*telego.Message, error) {
	msg := telego.SendVoiceParams{
		ChatID:  c.GetChatID(),
		Voice:   voice.Data,
		Caption: voice.Caption,
	}
	return c.Api.SendVoice(c.Ctx, &msg)
}

// SendVideoNote 发送圆形视频
func (c *Context) SendVideoNote(videoNote message.Resource) (*telego.Message, error) {
	msg := telego.SendVideoNoteParams{
		ChatID:    c.GetChatID(),
		VideoNote: videoNote.Data,
	}
	return c.Api.SendVideoNote(c.Ctx, &msg)
}

// SendSticker 发送贴纸
func (c *Context) SendSticker(sticker message.Resource) (*telego.Message, error) {
	msg := telego.SendStickerParams{
		ChatID:  c.GetChatID(),
		Sticker: sticker.Data,
	}
	return c.Api.SendSticker(c.Ctx, &msg)
}

// SendMediaGroup 发送媒体组
func (c *Context) SendMediaGroup(resources []message.Resource, caption string) ([]telego.Message, error) {
	if len(resources) == 0 {
		return nil, fmt.Errorf("资源列表为空")
	}

	// Telegram Media Group 限制最多 10 张
	if len(resources) > 10 {
		resources = resources[:10]
	}

	var mediaGroup []telego.InputMedia

	for i, res := range resources {
		media := &telego.InputMediaPhoto{
			Type:  telego.MediaTypePhoto,
			Media: res.Data,
		}

		// Telegram 限制：只有第一张图片可以有 caption
		if i == 0 && caption != "" {
			media.Caption = caption
		}

		mediaGroup = append(mediaGroup, media)
	}

	msg := telego.SendMediaGroupParams{
		ChatID: c.GetChatID(),
		Media:  mediaGroup,
	}

	return c.Api.SendMediaGroup(c.Ctx, &msg)
}

// SendMediaGroupFromPaths 从路径发送媒体组
func (c *Context) SendMediaGroupFromPaths(paths []string, caption string) ([]telego.Message, error) {
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
func (c *Context) SendLocation(latitude, longitude float64) (*telego.Message, error) {
	msg := telego.SendLocationParams{
		ChatID:    c.GetChatID(),
		Latitude:  latitude,
		Longitude: longitude,
	}
	return c.Api.SendLocation(c.Ctx, &msg)
}

// SendContact 发送联系人
func (c *Context) SendContact(phoneNumber, firstName string) (*telego.Message, error) {
	msg := telego.SendContactParams{
		ChatID:      c.GetChatID(),
		PhoneNumber: phoneNumber,
		FirstName:   firstName,
	}
	return c.Api.SendContact(c.Ctx, &msg)
}

// SendDice 发送骰子
func (c *Context) SendDice(emoji string) (*telego.Message, error) {
	msg := telego.SendDiceParams{
		ChatID: c.GetChatID(),
	}
	if emoji != "" {
		msg.Emoji = emoji
	}
	return c.Api.SendDice(c.Ctx, &msg)
}

// ============ 特殊操作方法 ============

// SendChatAction 发送聊天动作（如 typing, upload_photo 等）
func (c *Context) SendChatAction(action string) error {
	msg := telego.SendChatActionParams{
		ChatID: c.GetChatID(),
		Action: action,
	}
	return c.Api.SendChatAction(c.Ctx, &msg)
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

// SendVenue 发送地点
func (c *Context) SendVenue(latitude, longitude float64, title, address string) (*telego.Message, error) {
	msg := telego.SendVenueParams{
		ChatID:    c.GetChatID(),
		Latitude:  latitude,
		Longitude: longitude,
		Title:     title,
		Address:   address,
	}
	return c.Api.SendVenue(c.Ctx, &msg)
}
