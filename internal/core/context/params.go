package context

import (
	"strings"

	"github.com/mymmrac/telego"
)

// ============ 消息获取方法 ============

// GetMessage 获取消息，优先顺序：普通>普通编辑>频道>频道编辑
func (c *Context) GetMessage() *telego.Message {
	if c.Update.Message != nil {
		return c.Update.Message
	}
	if c.Update.EditedMessage != nil {
		return c.Update.EditedMessage
	}
	if c.Update.ChannelPost != nil {
		return c.Update.ChannelPost
	}
	if c.Update.EditedChannelPost != nil {
		return c.Update.EditedChannelPost
	}
	if c.Update.CallbackQuery != nil && c.Update.CallbackQuery.Message != nil {
		return c.Update.CallbackQuery.Message.Message()
	}
	return nil
}

// 获取bot id
func (c *Context) GetBotID() (int64, error) {
	bot, err := c.Api.GetMe(c.Ctx)
	if err != nil {
		return 0, err
	}
	return bot.ID, nil
}

// GetMessageID 获取消息 ID
func (c *Context) GetMessageID() int {
	msg := c.GetMessage()
	if msg != nil {
		return msg.MessageID
	}
	return 0
}

// GetMessageText 获取消息文本
func (c *Context) GetMessageText() string {
	msg := c.GetMessage()
	if msg != nil {
		return msg.Text
	}
	return ""
}

// GetCaption 获取媒体说明文本
func (c *Context) GetCaption() string {
	msg := c.GetMessage()
	if msg != nil {
		return msg.Caption
	}
	return ""
}

// GetTextOrCaption 获取文本或说明文本
func (c *Context) GetTextOrCaption() string {
	msg := c.GetMessage()
	if msg != nil {
		if msg.Text != "" {
			return msg.Text
		}
		return msg.Caption
	}
	return ""
}

// ============ 聊天相关方法 ============

// GetChatID 获取当前聊天 ID
func (c *Context) GetChatID() telego.ChatID {
	chat := c.GetChat()
	return chat.ChatID()
}

// GetChat 获取聊天对象
func (c *Context) GetChat() (chat telego.Chat) {
	if c.Update.Message != nil {
		return c.Update.Message.Chat
	}
	if c.Update.EditedMessage != nil {
		return c.Update.EditedMessage.Chat
	}
	if c.Update.ChannelPost != nil {
		return c.Update.ChannelPost.Chat
	}
	if c.Update.EditedChannelPost != nil {
		return c.Update.EditedChannelPost.Chat
	}
	if c.Update.CallbackQuery != nil && c.Update.CallbackQuery.Message != nil {
		return c.Update.CallbackQuery.Message.GetChat()
	}
	if c.Update.MyChatMember != nil {
		return c.Update.MyChatMember.Chat
	}
	if c.Update.ChatMember != nil {
		return c.Update.ChatMember.Chat
	}
	if c.Update.ChatJoinRequest != nil {
		return c.Update.ChatJoinRequest.Chat
	}
	return telego.Chat{
		ID:        0,
		Type:      "unknown",
		Title:     "unknown",
		Username:  "unknown",
		FirstName: "unknown",
		LastName:  "unknown",
	}
}

// GetChatTitle 获取当前聊天标题
func (c *Context) GetChatTitle() string {
	chat := c.GetChat()
	return chat.Title

}

// GetChatType 获取聊天类型 (private, group, supergroup, channel)
func (c *Context) GetChatType() string {
	chat := c.GetChat()
	return chat.Type

}

// ============ 用户相关方法 ============

// GetUser 获取用户对象
func (c *Context) GetUser() *telego.User {
	if c.Update.Message != nil && c.Update.Message.From != nil {
		return c.Update.Message.From
	}
	if c.Update.EditedMessage != nil && c.Update.EditedMessage.From != nil {
		return c.Update.EditedMessage.From
	}
	if c.Update.CallbackQuery != nil {
		return &c.Update.CallbackQuery.From
	}
	if c.Update.InlineQuery != nil {
		return &c.Update.InlineQuery.From
	}
	if c.Update.ChosenInlineResult != nil {
		return &c.Update.ChosenInlineResult.From
	}
	if c.Update.ShippingQuery != nil {
		return &c.Update.ShippingQuery.From
	}
	if c.Update.PreCheckoutQuery != nil {
		return &c.Update.PreCheckoutQuery.From
	}
	if c.Update.MyChatMember != nil {
		return &c.Update.MyChatMember.From
	}
	if c.Update.ChatMember != nil {
		return &c.Update.ChatMember.From
	}
	if c.Update.ChatJoinRequest != nil {
		return &c.Update.ChatJoinRequest.From
	}
	return nil
}

// GetUserID 获取当前用户 ID
func (c *Context) GetUserID() int64 {
	user := c.GetUser()
	if user != nil {
		return user.ID
	}
	return 0
}

// GetUsername 获取用户名
func (c *Context) GetUsername() string {
	user := c.GetUser()
	if user != nil {
		return user.Username
	}
	return ""
}

// GetFirstName 获取用户名字
func (c *Context) GetFirstName() string {
	user := c.GetUser()
	if user != nil {
		return user.FirstName
	}
	return ""
}

// GetLastName 获取用户姓氏
func (c *Context) GetLastName() string {
	user := c.GetUser()
	if user != nil {
		return user.LastName
	}
	return ""
}

// GetFullName 获取用户全名
func (c *Context) GetFullName() string {
	user := c.GetUser()
	if user != nil {
		fullName := user.FirstName
		if user.LastName != "" {
			fullName += " " + user.LastName
		}
		return fullName
	}
	return ""
}

// GetLanguageCode 获取用户语言代码
func (c *Context) GetLanguageCode() string {
	user := c.GetUser()
	if user != nil {
		return user.LanguageCode
	}
	return ""
}

// ============ 命令相关方法 ============

// GetCommand 获取命令（不包含 /）
func (c *Context) GetCommand() string {
	msg := c.GetMessage()
	if msg == nil || !c.IsCommand() {
		return ""
	}

	for _, entity := range msg.Entities {
		if entity.Type == "bot_command" && entity.Offset == 0 {
			// 命令文本在 msg.Text 里，从 Offset 开始，长度为 entity.Length
			cmd := msg.Text[entity.Offset : entity.Offset+entity.Length]
			// 去掉前面的 "/"
			if len(cmd) > 0 && cmd[0] == '/' {
				// 还要去掉 bot username（/start@mybot）
				for i, ch := range cmd {
					if ch == '@' {
						return cmd[1:i] // 去掉 '/' 和 '@...'
					}
				}
				return cmd[1:] // 直接去掉 '/'
			}
		}
	}

	return ""
}

// GetCommandArgs 获取命令参数
func (c *Context) GetCommandArgs() string {
	msg := c.GetMessage()
	if msg == nil || !c.IsCommand() {
		return ""
	}

	for _, entity := range msg.Entities {
		if entity.Type == "bot_command" && entity.Offset == 0 {
			if len(msg.Text) > entity.Length {
				return strings.TrimSpace(msg.Text[entity.Length:])
			}
			break
		}
	}

	return ""
}

// ============ 回调查询方法 ============

// GetCallbackQuery 获取回调查询
func (c *Context) GetCallbackQuery() *telego.CallbackQuery {
	return c.Update.CallbackQuery
}

// GetCallbackData 获取回调数据
func (c *Context) GetCallbackData() string {
	if c.Update.CallbackQuery != nil {
		return c.Update.CallbackQuery.Data
	}
	return ""
}

// GetCallbackQueryID 获取回调查询 ID
func (c *Context) GetCallbackQueryID() string {
	if c.Update.CallbackQuery != nil {
		return c.Update.CallbackQuery.ID
	}
	return ""
}

// ============ 内联查询方法 ============

// 获取内联查询
func (c *Context) GetInlineQuery() *telego.InlineQuery {
	return c.Update.InlineQuery
}

// 获取内联查询文本
func (c *Context) GetInlineQueryText() string {
	if c.Update.InlineQuery != nil {
		return c.Update.InlineQuery.Query
	}
	return ""
}

// 获取内联查询 ID
func (c *Context) GetInlineQueryID() string {
	if c.Update.InlineQuery != nil {
		return c.Update.InlineQuery.ID
	}
	return ""
}

// GetChosenInlineResult 获取选中的内联结果
func (c *Context) GetChosenInlineResult() *telego.ChosenInlineResult {
	return c.Update.ChosenInlineResult
}

// ============ 媒体相关方法 ============

// collectMedias 内部工具函数：收集单条消息的媒体（取最高质量）
func collectMedias(msg *telego.Message) (photos []string, videos []string, animations []string, documents []string, audios []string, voices []string, videoNotes []string) {
	if msg == nil {
		return
	}

	// Photo: Telegram 从小到大排序，取最后一张（最高质量）
	if len(msg.Photo) > 0 {
		best := msg.Photo[len(msg.Photo)-1].FileID
		photos = append(photos, best)
	}

	// Video
	if msg.Video != nil {
		videos = append(videos, msg.Video.FileID)
	}

	// Animation (GIF)
	if msg.Animation != nil {
		animations = append(animations, msg.Animation.FileID)
	}

	// Document
	if msg.Document != nil {
		if strings.HasPrefix(msg.Document.MimeType, "image/") {
			photos = append(photos, msg.Document.FileID)
		} else if strings.HasPrefix(msg.Document.MimeType, "video/") {
			videos = append(videos, msg.Document.FileID)
		} else {
			documents = append(documents, msg.Document.FileID)
		}
	}

	// Audio
	if msg.Audio != nil {
		audios = append(audios, msg.Audio.FileID)
	}

	// Voice
	if msg.Voice != nil {
		voices = append(voices, msg.Voice.FileID)
	}

	// VideoNote (圆形视频)
	if msg.VideoNote != nil {
		videoNotes = append(videoNotes, msg.VideoNote.FileID)
	}

	return
}

// GetMedias 获取所有媒体（包括回复消息）
func (c *Context) GetMedias() ([]string, bool) {
	var medias []string

	process := func(msg *telego.Message) {
		p, v, a, d, au, vo, vn := collectMedias(msg)
		medias = append(medias, p...)
		medias = append(medias, v...)
		medias = append(medias, a...)
		medias = append(medias, d...)
		medias = append(medias, au...)
		medias = append(medias, vo...)
		medias = append(medias, vn...)
	}

	process(c.Update.Message)
	if c.Update.Message != nil && c.Update.Message.ReplyToMessage != nil {
		process(c.Update.Message.ReplyToMessage)
	}

	if len(medias) == 0 {
		return nil, false
	}
	return medias, true
}

// GetPhotos 获取所有图片（包括回复消息）
func (c *Context) GetPhotos() ([]string, bool) {
	var photos []string

	process := func(msg *telego.Message) {
		p, _, _, _, _, _, _ := collectMedias(msg)
		photos = append(photos, p...)
	}

	process(c.Update.Message)
	if c.Update.Message != nil && c.Update.Message.ReplyToMessage != nil {
		process(c.Update.Message.ReplyToMessage)
	}

	if len(photos) == 0 {
		return nil, false
	}
	return photos, true
}

// GetPhoto 获取单张图片（包括回复消息）
func (c *Context) GetPhoto() (string, bool) {
	photos, ok := c.GetPhotos()
	if !ok || len(photos) == 0 {
		return "", false
	}
	return photos[0], true
}

// GetVideos 获取所有视频（包括回复消息）
func (c *Context) GetVideos() ([]string, bool) {
	var videos []string

	process := func(msg *telego.Message) {
		_, v, _, _, _, _, _ := collectMedias(msg)
		videos = append(videos, v...)
	}

	process(c.Update.Message)
	if c.Update.Message != nil && c.Update.Message.ReplyToMessage != nil {
		process(c.Update.Message.ReplyToMessage)
	}

	if len(videos) == 0 {
		return nil, false
	}
	return videos, true
}

// GetVideo 获取单个视频
func (c *Context) GetVideo() (string, bool) {
	videos, ok := c.GetVideos()
	if !ok || len(videos) == 0 {
		return "", false
	}
	return videos[0], true
}

// GetAnimations 获取所有动画/GIF
func (c *Context) GetAnimations() ([]string, bool) {
	var animations []string

	process := func(msg *telego.Message) {
		_, _, a, _, _, _, _ := collectMedias(msg)
		animations = append(animations, a...)
	}

	process(c.Update.Message)
	if c.Update.Message != nil && c.Update.Message.ReplyToMessage != nil {
		process(c.Update.Message.ReplyToMessage)
	}

	if len(animations) == 0 {
		return nil, false
	}
	return animations, true
}

// GetAnimation 获取单个动画/GIF
func (c *Context) GetAnimation() (string, bool) {
	animations, ok := c.GetAnimations()
	if !ok || len(animations) == 0 {
		return "", false
	}
	return animations[0], true
}

// GetDocuments 获取所有文档
func (c *Context) GetDocuments() ([]string, bool) {
	var documents []string

	process := func(msg *telego.Message) {
		_, _, _, d, _, _, _ := collectMedias(msg)
		documents = append(documents, d...)
	}

	process(c.Update.Message)
	if c.Update.Message != nil && c.Update.Message.ReplyToMessage != nil {
		process(c.Update.Message.ReplyToMessage)
	}

	if len(documents) == 0 {
		return nil, false
	}
	return documents, true
}

// GetDocument 获取单个文档
func (c *Context) GetDocument() (string, bool) {
	documents, ok := c.GetDocuments()
	if !ok || len(documents) == 0 {
		return "", false
	}
	return documents[0], true
}

// GetAudios 获取所有音频
func (c *Context) GetAudios() ([]string, bool) {
	var audios []string

	process := func(msg *telego.Message) {
		_, _, _, _, au, _, _ := collectMedias(msg)
		audios = append(audios, au...)
	}

	process(c.Update.Message)
	if c.Update.Message != nil && c.Update.Message.ReplyToMessage != nil {
		process(c.Update.Message.ReplyToMessage)
	}

	if len(audios) == 0 {
		return nil, false
	}
	return audios, true
}

// GetAudio 获取单个音频
func (c *Context) GetAudio() (string, bool) {
	audios, ok := c.GetAudios()
	if !ok || len(audios) == 0 {
		return "", false
	}
	return audios[0], true
}

// GetVoices 获取所有语音消息
func (c *Context) GetVoices() ([]string, bool) {
	var voices []string

	process := func(msg *telego.Message) {
		_, _, _, _, _, vo, _ := collectMedias(msg)
		voices = append(voices, vo...)
	}

	process(c.Update.Message)
	if c.Update.Message != nil && c.Update.Message.ReplyToMessage != nil {
		process(c.Update.Message.ReplyToMessage)
	}

	if len(voices) == 0 {
		return nil, false
	}
	return voices, true
}

// GetVoice 获取单个语音消息
func (c *Context) GetVoice() (string, bool) {
	voices, ok := c.GetVoices()
	if !ok || len(voices) == 0 {
		return "", false
	}
	return voices[0], true
}

// GetVideoNotes 获取所有圆形视频
func (c *Context) GetVideoNotes() ([]string, bool) {
	var videoNotes []string

	process := func(msg *telego.Message) {
		_, _, _, _, _, _, vn := collectMedias(msg)
		videoNotes = append(videoNotes, vn...)
	}

	process(c.Update.Message)
	if c.Update.Message != nil && c.Update.Message.ReplyToMessage != nil {
		process(c.Update.Message.ReplyToMessage)
	}

	if len(videoNotes) == 0 {
		return nil, false
	}
	return videoNotes, true
}

// GetVideoNote 获取单个圆形视频
func (c *Context) GetVideoNote() (string, bool) {
	videoNotes, ok := c.GetVideoNotes()
	if !ok || len(videoNotes) == 0 {
		return "", false
	}
	return videoNotes[0], true
}

// GetSticker 获取贴纸
func (c *Context) GetSticker() *telego.Sticker {
	if c.Update.Message != nil && c.Update.Message.Sticker != nil {
		return c.Update.Message.Sticker
	}
	return nil
}

// HasMedia 判断是否包含媒体
func (c *Context) HasMedia() bool {
	msg := c.GetMessage()
	if msg == nil {
		return false
	}
	return len(msg.Photo) > 0 || msg.Video != nil || msg.Animation != nil ||
		msg.Document != nil || msg.Audio != nil || msg.Voice != nil ||
		msg.VideoNote != nil || msg.Sticker != nil
}

// ============ 回复消息方法 ============

// GetReplyToMessage 获取被回复的消息
func (c *Context) GetReplyToMessage() *telego.Message {
	if c.Update.Message != nil {
		return c.Update.Message.ReplyToMessage
	}
	return nil
}

// GetReplyToMessageID 获取被回复消息的 ID
func (c *Context) GetReplyToMessageID() int {
	replyMsg := c.GetReplyToMessage()
	if replyMsg != nil {
		return replyMsg.MessageID
	}
	return 0
}

// ============ 位置和联系人方法 ============

// GetLocation 获取位置信息
func (c *Context) GetLocation() *telego.Location {
	if c.Update.Message != nil && c.Update.Message.Location != nil {
		return c.Update.Message.Location
	}
	if c.Update.EditedMessage != nil && c.Update.EditedMessage.Location != nil {
		return c.Update.EditedMessage.Location
	}
	return nil
}

// GetContact 获取联系人信息
func (c *Context) GetContact() *telego.Contact {
	if c.Update.Message != nil && c.Update.Message.Contact != nil {
		return c.Update.Message.Contact
	}
	return nil
}

// GetVenue 获取地点信息
func (c *Context) GetVenue() *telego.Venue {
	if c.Update.Message != nil && c.Update.Message.Venue != nil {
		return c.Update.Message.Venue
	}
	return nil
}

// GetPoll 获取投票信息
func (c *Context) GetPoll() *telego.Poll {
	if c.Update.Message != nil && c.Update.Message.Poll != nil {
		return c.Update.Message.Poll
	}
	if c.Update.Poll != nil {
		return c.Update.Poll
	}
	return nil
}

// GetPollAnswer 获取投票答案
func (c *Context) GetPollAnswer() *telego.PollAnswer {
	return c.Update.PollAnswer
}

// GetDice 获取骰子信息
func (c *Context) GetDice() *telego.Dice {
	if c.Update.Message != nil && c.Update.Message.Dice != nil {
		return c.Update.Message.Dice
	}
	return nil
}

// ============ 群组管理方法 ============

// GetNewChatMembers 获取新加入的成员
func (c *Context) GetNewChatMembers() []telego.User {
	if c.Update.Message != nil && c.Update.Message.NewChatMembers != nil {
		return c.Update.Message.NewChatMembers
	}
	return nil
}

// GetLeftChatMember 获取离开的成员
func (c *Context) GetLeftChatMember() *telego.User {
	if c.Update.Message != nil && c.Update.Message.LeftChatMember != nil {
		return c.Update.Message.LeftChatMember
	}
	return nil
}

// GetNewChatTitle 获取新的群组标题
func (c *Context) GetNewChatTitle() string {
	if c.Update.Message != nil {
		return c.Update.Message.NewChatTitle
	}
	return ""
}

// GetNewChatPhoto 获取新的群组头像
func (c *Context) GetNewChatPhoto() []telego.PhotoSize {
	if c.Update.Message != nil && c.Update.Message.NewChatPhoto != nil {
		return c.Update.Message.NewChatPhoto
	}
	return nil
}

// GetPinnedMessage 获取置顶的消息
func (c *Context) GetPinnedMessage() *telego.Message {
	if c.Update.Message != nil {
		return c.Update.Message.PinnedMessage.Message()
	}
	return nil
}

// GetMyChatMember 获取机器人自身的成员状态变化
func (c *Context) GetMyChatMember() *telego.ChatMemberUpdated {
	return c.Update.MyChatMember
}

// GetChatMember 获取群成员状态变化
func (c *Context) GetChatMember() *telego.ChatMemberUpdated {
	return c.Update.ChatMember
}

// GetChatJoinRequest 获取加群请求
func (c *Context) GetChatJoinRequest() *telego.ChatJoinRequest {
	return c.Update.ChatJoinRequest
}

// ============ 支付相关方法 ============

// GetShippingQuery 获取配送查询
func (c *Context) GetShippingQuery() *telego.ShippingQuery {
	return c.Update.ShippingQuery
}

// GetPreCheckoutQuery 获取预结账查询
func (c *Context) GetPreCheckoutQuery() *telego.PreCheckoutQuery {
	return c.Update.PreCheckoutQuery
}

// GetSuccessfulPayment 获取成功支付信息
func (c *Context) GetSuccessfulPayment() *telego.SuccessfulPayment {
	if c.Update.Message != nil && c.Update.Message.SuccessfulPayment != nil {
		return c.Update.Message.SuccessfulPayment
	}
	return nil
}

// ============ 其他实用方法 ============

// GetEntities 获取消息实体（如 @username, #hashtag, /command 等）
func (c *Context) GetEntities() []telego.MessageEntity {
	msg := c.GetMessage()
	if msg != nil {
		return msg.Entities
	}
	return nil
}

// GetCaptionEntities 获取媒体说明文本的实体
func (c *Context) GetCaptionEntities() []telego.MessageEntity {
	msg := c.GetMessage()
	if msg != nil {
		return msg.CaptionEntities
	}
	return nil
}

// HasText 是否包含文本
func (c *Context) HasText() bool {
	return c.GetMessageText() != ""
}

// GetUpdateID 获取更新 ID
func (c *Context) GetUpdateID() int {
	return c.Update.UpdateID
}
