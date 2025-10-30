package context

import (
	"fmt"
	"yueling_tg/internal/message"

	"github.com/mymmrac/telego"
)

// ============ 回复消息方法（Reply to Message）============

// Reply 回复当前消息
func (c *Context) Reply(text string) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID: c.GetChatID(),
		Text:   text,

		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// Replyf 格式化回复当前消息
func (c *Context) Replyf(format string, args ...interface{}) (*telego.Message, error) {
	return c.Reply(fmt.Sprintf(format, args...))
}

// ReplyMarkdown 使用 Markdown 格式回复
func (c *Context) ReplyMarkdown(text string) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID: c.GetChatID(),
		Text:   text,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
		ParseMode: telego.ModeMarkdown,
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// ReplyMarkdownV2 使用 MarkdownV2 格式回复
func (c *Context) ReplyMarkdownV2(text string) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID: c.GetChatID(),
		Text:   text,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
		ParseMode: telego.ModeMarkdownV2,
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// ReplyHTML 使用 HTML 格式回复
func (c *Context) ReplyHTML(text string) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID: c.GetChatID(),
		Text:   text,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
		ParseMode: telego.ModeHTML,
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// ReplyWithKeyboard 回复消息并带键盘
func (c *Context) ReplyWithKeyboard(text string, keyboard telego.ReplyMarkup) (*telego.Message, error) {
	msg := telego.SendMessageParams{
		ChatID: c.GetChatID(),
		Text:   text,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
		ReplyMarkup: keyboard,
	}
	return c.Api.SendMessage(c.Ctx, &msg)
}

// ReplyAudio 回复音频
func (c *Context) ReplyAudio(audio message.Resource) (*telego.Message, error) {
	msg := telego.SendAudioParams{
		ChatID:  c.GetChatID(),
		Audio:   audio.Data,
		Caption: audio.Caption,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
	}
	return c.Api.SendAudio(c.Ctx, &msg)
}

// ReplyVideo 回复视频
func (c *Context) ReplyVideo(video message.Resource) (*telego.Message, error) {
	msg := telego.SendVideoParams{
		ChatID:  c.GetChatID(),
		Video:   video.Data,
		Caption: video.Caption,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
	}
	return c.Api.SendVideo(c.Ctx, &msg)
}

// ReplyAnimation 回复动画/GIF
func (c *Context) ReplyAnimation(animation message.Resource) (*telego.Message, error) {
	msg := telego.SendAnimationParams{
		ChatID:    c.GetChatID(),
		Animation: animation.Data,
		Caption:   animation.Caption,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
	}
	return c.Api.SendAnimation(c.Ctx, &msg)
}

// ReplyDocument 回复文档
func (c *Context) ReplyDocument(document message.Resource) (*telego.Message, error) {
	msg := telego.SendDocumentParams{
		ChatID:   c.GetChatID(),
		Document: document.Data,
		Caption:  document.Caption,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
	}
	return c.Api.SendDocument(c.Ctx, &msg)
}

// ReplyVoice 回复语音消息
func (c *Context) ReplyVoice(voice message.Resource) (*telego.Message, error) {
	msg := telego.SendVoiceParams{
		ChatID:  c.GetChatID(),
		Voice:   voice.Data,
		Caption: voice.Caption,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
	}
	return c.Api.SendVoice(c.Ctx, &msg)
}

// ReplyVideoNote 回复圆形视频
func (c *Context) ReplyVideoNote(videoNote message.Resource) (*telego.Message, error) {
	msg := telego.SendVideoNoteParams{
		ChatID:    c.GetChatID(),
		VideoNote: videoNote.Data,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
	}
	return c.Api.SendVideoNote(c.Ctx, &msg)
}

// ReplySticker 回复贴纸
func (c *Context) ReplySticker(sticker message.Resource) (*telego.Message, error) {
	msg := telego.SendStickerParams{
		ChatID:  c.GetChatID(),
		Sticker: sticker.Data,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
	}
	return c.Api.SendSticker(c.Ctx, &msg)
}

// ReplyReaction 使用标准 emoji 回复
func (c *Context) ReplyReaction(emoji string) error {
	return c.replyReaction(&telego.ReactionTypeEmoji{
		Type:  "emoji",
		Emoji: emoji,
	}, false)
}

func (c *Context) ReplyReactionAck() error {
	return c.replyReaction(&telego.ReactionTypeEmoji{
		Type:  "emoji",
		Emoji: "👌",
	}, false)
}

// ReplyCustomReaction 使用自定义表情回复
func (c *Context) ReplyCustomReaction(customEmojiID string) error {
	return c.replyReaction(&telego.ReactionTypeCustomEmoji{
		Type:          "custom_emoji",
		CustomEmojiID: customEmojiID,
	}, false)
}

func (c *Context) replyReaction(reaction interface{}, isBig bool) error {
	var reactions []telego.ReactionType

	switch r := reaction.(type) {
	case telego.ReactionType:
		reactions = []telego.ReactionType{r}
	case []telego.ReactionType:
		reactions = r
	}

	params := &telego.SetMessageReactionParams{
		ChatID:    telego.ChatID{ID: c.GetChat().ID},
		MessageID: c.GetMessageID(),
		Reaction:  reactions,
		IsBig:     isBig,
	}
	return c.Api.SetMessageReaction(c.Ctx, params)
}

// ============ 媒体组发送方法 ============

// ReplyLocation 回复位置
func (c *Context) ReplyLocation(latitude, longitude float64) (*telego.Message, error) {
	msg := telego.SendLocationParams{
		ChatID:    c.GetChatID(),
		Latitude:  latitude,
		Longitude: longitude,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
	}
	return c.Api.SendLocation(c.Ctx, &msg)
}

// ReplyVenue 回复地点
func (c *Context) ReplyVenue(latitude, longitude float64, title, address string) (*telego.Message, error) {
	msg := telego.SendVenueParams{
		ChatID:    c.GetChatID(),
		Latitude:  latitude,
		Longitude: longitude,
		Title:     title,
		Address:   address,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: c.GetMessageID(),
		},
	}
	return c.Api.SendVenue(c.Ctx, &msg)
}
