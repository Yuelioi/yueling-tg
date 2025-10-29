package context

import "github.com/mymmrac/telego"

func (c *Context) GetBot() (*telego.User, error) {
	return c.Api.GetMe(c.Ctx)
}

func (c *Context) GetBotFullname() string {
	bot, err := c.GetBot()
	if err != nil {
		return "机器人"
	}
	return bot.FirstName + bot.LastName
}

// 获取Bot当前成员信息
func (c *Context) GetBotMember() (member telego.ChatMember, err error) {
	botUser, err := c.Api.GetMe(c.Ctx)
	if err != nil {
		return
	}

	return c.Api.GetChatMember(c.Ctx, &telego.GetChatMemberParams{
		ChatID: c.GetChatID(),
		UserID: botUser.ID,
	})

}

func (c *Context) IsBotAdmin() bool {
	member, err := c.GetBotMember()
	if err != nil {
		return false
	}

	switch member.(type) {
	case *telego.ChatMemberOwner, *telego.ChatMemberAdministrator:
		return true
	default:
		return false
	}
}

//判断 Bot 是否有限制成员的权限
func (c *Context) CanBotRestrictMembers() bool {
	member, err := c.GetBotMember()
	if err != nil {
		return false
	}

	switch m := member.(type) {
	case *telego.ChatMemberOwner:
		// 群主拥有所有权限
		return true
	case *telego.ChatMemberAdministrator:
		return m.CanRestrictMembers
	default:
		return false
	}
}

//判断 Bot 是否有撤回的权限
func (c *Context) CanBotDeleteMessage() bool {
	member, err := c.GetBotMember()
	if err != nil {
		return false
	}

	switch m := member.(type) {
	case *telego.ChatMemberOwner:
		// 群主拥有所有权限
		return true
	case *telego.ChatMemberAdministrator:
		return m.CanDeleteMessages
	default:
		return false
	}
}

func (c *Context) IsAdmin() bool {
	member, err := c.Api.GetChatMember(c.Ctx, &telego.GetChatMemberParams{
		ChatID: c.GetChatID(),
		UserID: c.GetUserID(),
	})
	if err != nil {
		return false
	}

	switch member.(type) {
	case *telego.ChatMemberOwner, *telego.ChatMemberAdministrator:
		return true

	default:
		return false
	}
}
