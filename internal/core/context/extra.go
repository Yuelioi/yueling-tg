package context

import (
	"fmt"
	"time"

	"github.com/mymmrac/telego"
)

// ============ 消息编辑方法 ============

// EditMessageText 编辑消息文本
func (c *Context) EditMessageText(messageID int, text string) (*telego.Message, error) {
	msg := telego.EditMessageTextParams{
		ChatID:    c.GetChatID(),
		MessageID: messageID,
		Text:      text,
	}
	return c.Api.EditMessageText(c.Ctx, &msg)
}

// EditMessageTextWithOptions 编辑消息文本（自定义选项）
func (c *Context) EditMessageTextWithOptions(messageID int, text string, options func(*telego.EditMessageTextParams)) (*telego.Message, error) {
	msg := telego.EditMessageTextParams{
		ChatID:    c.GetChatID(),
		MessageID: messageID,
		Text:      text,
	}
	if options != nil {
		options(&msg)
	}
	return c.Api.EditMessageText(c.Ctx, &msg)
}

// EditCurrentMessage 编辑当前消息（用于 CallbackQuery）
func (c *Context) EditCurrentMessage(text string) (*telego.Message, error) {
	if c.Update.CallbackQuery == nil || c.Update.CallbackQuery.Message == nil {
		return nil, fmt.Errorf("no callback query message to edit")
	}
	return c.EditMessageText(c.Update.CallbackQuery.Message.GetMessageID(), text)
}

// EditMessageCaption 编辑媒体说明文本
func (c *Context) EditMessageCaption(messageID int, caption string) (*telego.Message, error) {
	msg := telego.EditMessageCaptionParams{
		ChatID:    c.GetChatID(),
		MessageID: messageID,
		Caption:   caption,
	}
	return c.Api.EditMessageCaption(c.Ctx, &msg)
}

// EditMessageReplyMarkup 编辑消息的键盘
func (c *Context) EditMessageReplyMarkup(messageID int, keyboard *telego.InlineKeyboardMarkup) (*telego.Message, error) {
	msg := telego.EditMessageReplyMarkupParams{
		ChatID:      c.GetChatID(),
		MessageID:   messageID,
		ReplyMarkup: keyboard,
	}
	return c.Api.EditMessageReplyMarkup(c.Ctx, &msg)
}

// EditCurrentMessageWithKeyboard 编辑当前消息并更新键盘
func (c *Context) EditCurrentMessageWithKeyboard(text string, keyboard telego.InlineKeyboardMarkup) (*telego.Message, error) {
	if c.Update.CallbackQuery == nil || c.Update.CallbackQuery.Message == nil {
		return nil, fmt.Errorf("no callback query message to edit")
	}
	msg := telego.EditMessageTextParams{
		ChatID:      c.GetChatID(),
		MessageID:   c.Update.CallbackQuery.Message.GetMessageID(),
		Text:        text,
		ReplyMarkup: &keyboard,
	}
	return c.Api.EditMessageText(c.Ctx, &msg)
}

// EditMessageWithKeyboard 编辑消息并更新键盘
func (c *Context) EditMessageWithKeyboard(messageID int, text string, keyboard telego.InlineKeyboardMarkup) (*telego.Message, error) {
	msg := telego.EditMessageTextParams{
		ChatID:      c.GetChatID(),
		MessageID:   messageID,
		Text:        text,
		ReplyMarkup: &keyboard,
	}
	return c.Api.EditMessageText(c.Ctx, &msg)
}

// ============ 消息删除方法 ============

// DeleteMessage 删除消息
func (c *Context) DeleteMessage(messageID int) error {
	return c.Api.DeleteMessage(c.Ctx, &telego.DeleteMessageParams{
		ChatID:    c.GetChatID(),
		MessageID: messageID,
	})
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
	return c.Api.AnswerCallbackQuery(c.Ctx, &telego.AnswerCallbackQueryParams{
		CallbackQueryID: c.Update.CallbackQuery.ID,
		Text:            text,
	})
}

// AnswerCallbackWithAlert 回答回调查询并显示警告弹窗
func (c *Context) AnswerCallbackWithAlert(text string) error {
	if c.Update.CallbackQuery == nil {
		return fmt.Errorf("no callback query to answer")
	}
	showAlert := true
	return c.Api.AnswerCallbackQuery(c.Ctx, &telego.AnswerCallbackQueryParams{
		CallbackQueryID: c.Update.CallbackQuery.ID,
		Text:            text,
		ShowAlert:       showAlert,
	})
}

// AnswerCallbackWithURL 回答回调查询并打开 URL
func (c *Context) AnswerCallbackWithURL(url string) error {
	if c.Update.CallbackQuery == nil {
		return fmt.Errorf("no callback query to answer")
	}
	return c.Api.AnswerCallbackQuery(c.Ctx, &telego.AnswerCallbackQueryParams{
		CallbackQueryID: c.Update.CallbackQuery.ID,
		URL:             url,
	})
}

// ============ 内联查询响应方法 ============

// AnswerInlineQuery 回答内联查询
func (c *Context) AnswerInlineQuery(results []telego.InlineQueryResult) error {
	if c.Update.InlineQuery == nil {
		return fmt.Errorf("no inline query to answer")
	}
	return c.Api.AnswerInlineQuery(c.Ctx, &telego.AnswerInlineQueryParams{
		InlineQueryID: c.Update.InlineQuery.ID,
		Results:       results,
	})
}

// AnswerInlineQueryWithOptions 回答内联查询(自定义选项)
func (c *Context) AnswerInlineQueryWithOptions(results []telego.InlineQueryResult, options func(*telego.AnswerInlineQueryParams)) error {
	if c.Update.InlineQuery == nil {
		return fmt.Errorf("no inline query to answer")
	}
	params := &telego.AnswerInlineQueryParams{
		InlineQueryID: c.Update.InlineQuery.ID,
		Results:       results,
	}
	if options != nil {
		options(params)
	}
	return c.Api.AnswerInlineQuery(c.Ctx, params)
}

// ============ 转发方法 ============

// ForwardMessage 转发消息
func (c *Context) ForwardMessage(toChatID int64, fromChatID int64, messageID int) (*telego.Message, error) {
	return c.Api.ForwardMessage(c.Ctx, &telego.ForwardMessageParams{
		ChatID:     telego.ChatID{ID: toChatID},
		FromChatID: telego.ChatID{ID: fromChatID},
		MessageID:  messageID,
	})
}

// ForwardCurrentMessage 转发当前消息到指定聊天
func (c *Context) ForwardCurrentMessage(toChatID int64) (*telego.Message, error) {
	chatID := c.GetChatID()
	// 将 ChatID 转换为 int64

	return c.ForwardMessage(toChatID, chatID.ID, c.GetMessageID())
}

// CopyMessage 复制消息(不显示转发来源)
func (c *Context) CopyMessage(toChatID int64, fromChatID int64, messageID int) (*telego.MessageID, error) {
	return c.Api.CopyMessage(c.Ctx, &telego.CopyMessageParams{
		ChatID:     telego.ChatID{ID: toChatID},
		FromChatID: telego.ChatID{ID: fromChatID},
		MessageID:  messageID,
	})
}

// CopyCurrentMessage 复制当前消息到指定聊天
func (c *Context) CopyCurrentMessage(toChatID int64) (*telego.MessageID, error) {
	chatID := c.GetChatID()
	return c.CopyMessage(toChatID, chatID.ID, c.GetMessageID())
}

// ============ 群组管理方法 ============

// PinMessage 置顶消息
func (c *Context) PinMessage(messageID int, disableNotification bool) error {
	return c.Api.PinChatMessage(c.Ctx, &telego.PinChatMessageParams{
		ChatID:              c.GetChatID(),
		MessageID:           messageID,
		DisableNotification: disableNotification,
	})
}

// PinCurrentMessage 置顶当前消息
func (c *Context) PinCurrentMessage(disableNotification bool) error {
	return c.PinMessage(c.GetMessageID(), disableNotification)
}

// UnpinMessage 取消置顶消息
func (c *Context) UnpinMessage(messageID int) error {
	return c.Api.UnpinChatMessage(c.Ctx, &telego.UnpinChatMessageParams{
		ChatID:    c.GetChatID(),
		MessageID: messageID,
	})
}

// UnpinAllMessages 取消置顶所有消息
func (c *Context) UnpinAllMessages() error {
	return c.Api.UnpinAllChatMessages(c.Ctx, &telego.UnpinAllChatMessagesParams{
		ChatID: c.GetChatID(),
	})
}

// BanChatMember 封禁群成员
func (c *Context) BanChatMember(userID int64) error {
	return c.Api.BanChatMember(c.Ctx, &telego.BanChatMemberParams{
		ChatID: c.GetChatID(),
		UserID: userID,
	})
}

// UnbanChatMember 解封群成员
func (c *Context) UnbanChatMember(userID int64) error {
	return c.Api.UnbanChatMember(c.Ctx, &telego.UnbanChatMemberParams{
		ChatID: c.GetChatID(),
		UserID: userID,
	})
}

// RestrictChatMember 限制群成员
func (c *Context) RestrictChatMember(userID int64, permissions telego.ChatPermissions) error {
	return c.Api.RestrictChatMember(c.Ctx, &telego.RestrictChatMemberParams{
		ChatID:      c.GetChatID(),
		UserID:      userID,
		Permissions: permissions,
	})
}

// PromoteChatMember 提升群成员为管理员
func (c *Context) PromoteChatMember(userID int64) error {
	return c.Api.PromoteChatMember(c.Ctx, &telego.PromoteChatMemberParams{
		ChatID: c.GetChatID(),
		UserID: userID,
	})
}

// LeaveChat 离开群组/频道
func (c *Context) LeaveChat() error {
	return c.Api.LeaveChat(c.Ctx, &telego.LeaveChatParams{
		ChatID: c.GetChatID(),
	})
}

func (c *Context) MuteUser(userID int64, duration time.Duration) error {
	until := time.Now().Add(duration).Unix()

	params := &telego.RestrictChatMemberParams{
		ChatID: c.GetChatID(),
		UserID: userID,
		Permissions: telego.ChatPermissions{
			CanSendMessages:       boolPtr(false),
			CanSendAudios:         boolPtr(false),
			CanSendDocuments:      boolPtr(false),
			CanSendPhotos:         boolPtr(false),
			CanSendVideos:         boolPtr(false),
			CanSendVideoNotes:     boolPtr(false),
			CanSendVoiceNotes:     boolPtr(false),
			CanSendPolls:          boolPtr(false),
			CanAddWebPagePreviews: boolPtr(false),
			CanChangeInfo:         boolPtr(false),
			CanInviteUsers:        boolPtr(false),
			CanPinMessages:        boolPtr(false),
		},
		UntilDate: until,
	}

	return c.Api.RestrictChatMember(c.Ctx, params)
}

func (c *Context) GetFile(fileID string) (*telego.File, error) {
	return c.Api.GetFile(c.Ctx, &telego.GetFileParams{
		FileID: fileID,
	})
}
func (c *Context) GetFileDirectURL(fileID string) (string, error) {
	file, err := c.GetFile(fileID)
	if err != nil {
		return "", err
	}
	return file.FilePath, nil
}
