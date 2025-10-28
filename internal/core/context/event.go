package context

//
// ────────────────────────────────
// 📩 Ⅰ. 消息类事件（用户或频道主动发出的消息）
// ────────────────────────────────
//

// IsChatMessage 判断是否为普通聊天消息（私聊、群聊、频道）
// 包括文本、图片、语音、视频、文件等类型
func (c *Context) IsChatMessage() bool {
	return c.GetMessage() != nil
}

// IsEditedMessage 判断是否为用户编辑过的消息（私聊/群聊）
func (c *Context) IsEditedMessage() bool {
	return c.Update.EditedMessage != nil
}

// IsChannelMessage 判断是否为频道发出的原始消息
func (c *Context) IsChannelMessage() bool {
	return c.Update.ChannelPost != nil
}

// IsEditedChannelMessage 判断是否为编辑后的频道消息
func (c *Context) IsEditedChannelMessage() bool {
	return c.Update.EditedChannelPost != nil
}

// ✅ 综合判断：是否为消息类事件
func (c *Context) IsMessage() bool {
	return c.IsChatMessage() ||
		c.IsEditedMessage() ||
		c.IsChannelMessage() ||
		c.IsEditedChannelMessage()
}

//
// ────────────────────────────────
// 🔔 Ⅱ. 通知类事件（系统、状态变化）
// ────────────────────────────────
//

// IsChatMemberUpdate 判断是否为群成员状态变化事件
// 如：用户被禁言、被踢出、被提权等
func (c *Context) IsChatMemberUpdate() bool {
	return c.Update.ChatMember != nil
}

// IsMyChatMemberUpdate 判断是否为机器人自身状态变化事件
// 如：机器人被加入、被移除群组或被封禁
func (c *Context) IsMyChatMemberUpdate() bool {
	return c.Update.MyChatMember != nil
}

// IsPoll 判断是否为投票或答题事件
// 包含群组投票创建（Poll）和用户回答（PollAnswer）
func (c *Context) IsPoll() bool {
	return c.Update.Poll != nil || c.Update.PollAnswer != nil
}

// IsNewMember 判断是否有新成员加入群聊（普通群/超级群）
func (c *Context) IsNewMember() bool {
	msg := c.GetMessage()
	return msg != nil && msg.NewChatMembers != nil
}

// IsLeftMember 判断是否有成员退出或被移出群聊
func (c *Context) IsLeftMember() bool {
	msg := c.GetMessage()
	return msg != nil && msg.LeftChatMember != nil
}

// ✅ 综合判断：是否为通知类事件
func (c *Context) IsNotice() bool {
	return c.IsChatMemberUpdate() ||
		c.IsMyChatMemberUpdate() ||
		c.IsPoll() ||
		c.IsNewMember() ||
		c.IsLeftMember()
}

//
// ────────────────────────────────
// 🧩 Ⅲ. 回调类事件（交互式事件，如按钮、内联查询、请求）
// ────────────────────────────────
//

// IsCallbackQuery 判断是否为按钮回调事件（InlineKeyboardButton）
// 当用户点击 inline 按钮后触发
func (c *Context) IsCallbackQuery() bool {
	return c.Update.CallbackQuery != nil
}

// IsInlineQuery 判断是否为内联查询事件
// 当用户在输入框中以 @BotUsername 开头进行 inline 查询时触发
func (c *Context) IsInlineQuery() bool {
	return c.Update.InlineQuery != nil
}

// IsChosenInlineResult 判断是否为内联查询结果被选中事件
// 用户从 inline 查询结果中选择一项后触发
func (c *Context) IsChosenInlineResult() bool {
	return c.Update.ChosenInlineResult != nil
}

// IsJoinRequest 判断是否为入群申请事件（需审批的群/频道）
// 当用户请求加入私有群/频道时触发
func (c *Context) IsJoinRequest() bool {
	return c.Update.ChatJoinRequest != nil
}

// ✅ 综合判断：是否为回调类事件
func (c *Context) IsCallback() bool {
	return c.IsCallbackQuery() ||
		c.IsInlineQuery() ||
		c.IsChosenInlineResult() ||
		c.IsJoinRequest()
}

//
// ────────────────────────────────
// 🧠 其他辅助判断
// ────────────────────────────────
//

// IsCommand 判断是否为命令消息（以 “/” 开头的 Bot 指令）
func (c *Context) IsCommand() bool {
	msg := c.GetMessage()
	return msg != nil && msg.IsCommand()
}

// IsPrivate 判断是否为私聊
func (c *Context) IsPrivate() bool {
	msg := c.GetMessage()
	return msg != nil && msg.Chat.IsPrivate()
}

// IsGroup 判断是否为普通群聊
func (c *Context) IsGroup() bool {
	msg := c.GetMessage()
	return msg != nil && msg.Chat.IsGroup()
}

// IsSuperGroup 判断是否为超级群聊
func (c *Context) IsSuperGroup() bool {
	msg := c.GetMessage()
	return msg != nil && msg.Chat.IsSuperGroup()
}
