package admin

import (
	"fmt"
	"strings"

	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/dsl/permission"
	"yueling_tg/pkg/plugin/params"

	"github.com/mymmrac/telego"
)

var _ plugin.Plugin = (*AdminPlugin)(nil)

// -------------------- 插件结构 --------------------

type AdminPlugin struct {
	*plugin.Base
}

func New() plugin.Plugin {
	ap := &AdminPlugin{}

	info := &plugin.PluginInfo{
		ID:          "admin",
		Name:        "管理员管理",
		Description: "设置和管理群组管理员",
		Version:     "1.0.0",
		Author:      "月离",
		Usage:       "设置管理员（回复用户消息）/ 取消管理员（回复用户消息）/ 管理员列表",
		Group:       "管理",
	}

	builder := plugin.New().
		Info(info)

	// 命令处理 - 需要是群主或有权限的管理员才能使用
	builder.OnCommand("设置管理员").When(permission.GroupOwner()).Block(true).Do(ap.handlePromoteAdmin)
	builder.OnCommand("取消管理员").When(permission.GroupOwner()).Block(true).Do(ap.handleDemoteAdmin)
	builder.OnCommand("管理员列表").When(permission.GroupAdminOrOwner()).Do(ap.handleListAdmins)
	builder.OnCommand("禁言").When(permission.GroupAdminOrOwner()).Block(true).Do(ap.handleMute)
	builder.OnCommand("解除禁言").When(permission.GroupAdminOrOwner()).Block(true).Do(ap.handleUnmute)
	builder.OnCommand("踢出").When(permission.GroupAdminOrOwner()).Block(true).Do(ap.handleKick)

	return builder.Go(ap)
}

// -------------------- 命令处理 --------------------

// 设置管理员
func (ap *AdminPlugin) handlePromoteAdmin(c *context.Context, cmdCtx params.CommandContext) {
	if !c.IsGroup() && !c.IsSuperGroup() {
		c.Reply("❌ 此命令仅在群组中可用")
		return
	}

	msg := c.GetMessage()
	if msg == nil {
		return
	}

	// 获取目标用户
	targetUser := ap.getTargetUser(c, msg)
	if targetUser == nil {
		c.Reply("❌ 请回复要设置为管理员的用户消息，或使用 @用户名")
		return
	}

	t := true
	// 提升为管理员
	params := &telego.PromoteChatMemberParams{
		ChatID:             c.GetChatID(),
		UserID:             targetUser.ID,
		CanChangeInfo:      &t,
		CanDeleteMessages:  &t,
		CanInviteUsers:     &t,
		CanRestrictMembers: &t,
		CanPinMessages:     &t,
	}

	err := c.Api.PromoteChatMember(c.Ctx, params)
	if err != nil {
		ap.Log.Error().Err(err).Msg("设置管理员失败")
		c.Reply("❌ 设置管理员失败，还没有足够的权限哦~")
		return
	}

	fullName := targetUser.FirstName
	if targetUser.LastName != "" {
		fullName += " " + targetUser.LastName
	}

	ap.Log.Info().
		Int64("user_id", targetUser.ID).
		Str("username", targetUser.Username).
		Msg("设置为管理员")

	c.Replyf("✅ 已将 %s 设置为管理员", fullName)
}

// 取消管理员
func (ap *AdminPlugin) handleDemoteAdmin(c *context.Context, cmdCtx params.CommandContext) {
	if !c.IsGroup() && !c.IsSuperGroup() {
		c.Reply("❌ 此命令仅在群组中可用")
		return
	}

	msg := c.GetMessage()
	if msg == nil {
		return
	}

	// 获取目标用户
	targetUser := ap.getTargetUser(c, msg)
	if targetUser == nil {
		c.Reply("❌ 请回复要取消管理员的用户消息，或使用 @用户名")
		return
	}

	f := false
	// 取消管理员权限
	params := &telego.PromoteChatMemberParams{
		ChatID:             c.GetChatID(),
		UserID:             targetUser.ID,
		CanChangeInfo:      &f,
		CanDeleteMessages:  &f,
		CanInviteUsers:     &f,
		CanRestrictMembers: &f,
		CanPinMessages:     &f,
		CanPromoteMembers:  &f,
	}

	err := c.Api.PromoteChatMember(c.Ctx, params)
	if err != nil {
		ap.Log.Error().Err(err).Msg("取消管理员失败")
		c.Reply("❌ 取消管理员失败，请确保机器人有足够的权限")
		return
	}

	fullName := targetUser.FirstName
	if targetUser.LastName != "" {
		fullName += " " + targetUser.LastName
	}

	ap.Log.Info().
		Int64("user_id", targetUser.ID).
		Str("username", targetUser.Username).
		Msg("取消管理员")

	c.Replyf("✅ 已取消 %s 的管理员权限", fullName)
}

// 管理员列表
func (ap *AdminPlugin) handleListAdmins(c *context.Context, cmdCtx params.CommandContext) {
	if !c.IsGroup() && !c.IsSuperGroup() {
		c.Reply("❌ 此命令仅在群组中可用")
		return
	}

	params := &telego.GetChatAdministratorsParams{
		ChatID: c.GetChatID(),
	}

	admins, err := c.Api.GetChatAdministrators(c.Ctx, params)
	if err != nil {
		ap.Log.Error().Err(err).Msg("获取管理员列表失败")
		c.Reply("❌ 获取管理员列表失败")
		return
	}

	if len(admins) == 0 {
		c.Reply("当前没有管理员")
		return
	}

	var builder strings.Builder
	builder.WriteString("👥 当前管理员列表：\n\n")

	for i, admin := range admins {
		user := admin.MemberUser()
		fullName := user.FirstName
		if user.LastName != "" {
			fullName += " " + user.LastName
		}

		// 获取角色
		role := "管理员"
		switch member := admin.(type) {
		case *telego.ChatMemberOwner:
			role = "👑 群主"
		case *telego.ChatMemberAdministrator:
			if member.CustomTitle != "" {
				role = "👤 " + member.CustomTitle
			} else {
				role = "👤 管理员"
			}
		}

		builder.WriteString(fmt.Sprintf("%d. %s %s", i+1, role, fullName))
		if user.Username != "" {
			builder.WriteString(fmt.Sprintf(" (@%s)", user.Username))
		}
		builder.WriteString("\n")
	}

	c.Reply(builder.String())
}

// 禁言用户
func (ap *AdminPlugin) handleMute(c *context.Context, cmdCtx params.CommandContext) {
	if !c.IsGroup() && !c.IsSuperGroup() {
		c.Reply("❌ 此命令仅在群组中可用")
		return
	}

	msg := c.GetMessage()
	if msg == nil {
		return
	}

	// 获取目标用户
	targetUser := ap.getTargetUser(c, msg)
	if targetUser == nil {
		c.Reply("❌ 请回复要禁言的用户消息")
		return
	}

	f := false
	// 禁言（移除发送消息权限）
	permissions := telego.ChatPermissions{
		CanSendMessages:       &f,
		CanSendAudios:         &f,
		CanSendDocuments:      &f,
		CanSendPhotos:         &f,
		CanSendVideos:         &f,
		CanSendVideoNotes:     &f,
		CanSendVoiceNotes:     &f,
		CanSendPolls:          &f,
		CanSendOtherMessages:  &f,
		CanAddWebPagePreviews: &f,
	}

	params := &telego.RestrictChatMemberParams{
		ChatID:      c.GetChatID(),
		UserID:      targetUser.ID,
		Permissions: permissions,
	}

	err := c.Api.RestrictChatMember(c.Ctx, params)
	if err != nil {
		ap.Log.Error().Err(err).Msg("禁言失败")
		c.Reply("❌ 禁言失败，请确保机器人有足够的权限")
		return
	}

	fullName := targetUser.FirstName
	if targetUser.LastName != "" {
		fullName += " " + targetUser.LastName
	}

	ap.Log.Info().
		Int64("user_id", targetUser.ID).
		Str("username", targetUser.Username).
		Msg("禁言用户")

	c.Replyf("✅ 已禁言 %s", fullName)
}

// 解除禁言
func (ap *AdminPlugin) handleUnmute(c *context.Context, cmdCtx params.CommandContext) {
	if !c.IsGroup() && !c.IsSuperGroup() {
		c.Reply("❌ 此命令仅在群组中可用")
		return
	}

	msg := c.GetMessage()
	if msg == nil {
		return
	}

	// 获取目标用户
	targetUser := ap.getTargetUser(c, msg)
	if targetUser == nil {
		c.Reply("❌ 请回复要解除禁言的用户消息")
		return
	}
	t := true

	// 恢复发送消息权限
	permissions := telego.ChatPermissions{
		CanSendMessages:       &t,
		CanSendAudios:         &t,
		CanSendDocuments:      &t,
		CanSendPhotos:         &t,
		CanSendVideos:         &t,
		CanSendVideoNotes:     &t,
		CanSendVoiceNotes:     &t,
		CanSendPolls:          &t,
		CanSendOtherMessages:  &t,
		CanAddWebPagePreviews: &t,
	}

	params := &telego.RestrictChatMemberParams{
		ChatID:      c.GetChatID(),
		UserID:      targetUser.ID,
		Permissions: permissions,
	}

	err := c.Api.RestrictChatMember(c.Ctx, params)
	if err != nil {
		ap.Log.Error().Err(err).Msg("解除禁言失败")
		c.Reply("❌ 解除禁言失败，请确保机器人有足够的权限")
		return
	}

	fullName := targetUser.FirstName
	if targetUser.LastName != "" {
		fullName += " " + targetUser.LastName
	}

	ap.Log.Info().
		Int64("user_id", targetUser.ID).
		Str("username", targetUser.Username).
		Msg("解除禁言")

	c.Replyf("✅ 已解除 %s 的禁言", fullName)
}

// 踢出群组
func (ap *AdminPlugin) handleKick(c *context.Context, cmdCtx params.CommandContext) {
	if !c.IsGroup() && !c.IsSuperGroup() {
		c.Reply("❌ 此命令仅在群组中可用")
		return
	}

	msg := c.GetMessage()
	if msg == nil {
		return
	}

	// 获取目标用户
	targetUser := ap.getTargetUser(c, msg)
	if targetUser == nil {
		c.Reply("❌ 请回复要踢出的用户消息")
		return
	}

	// 踢出用户
	params := &telego.BanChatMemberParams{
		ChatID: c.GetChatID(),
		UserID: targetUser.ID,
	}

	err := c.Api.BanChatMember(c.Ctx, params)
	if err != nil {
		ap.Log.Error().Err(err).Msg("踢出失败")
		c.Reply("❌ 踢出失败，请确保机器人有足够的权限")
		return
	}

	// 立即解封（这样用户可以重新加入）
	unbanParams := &telego.UnbanChatMemberParams{
		ChatID: c.GetChatID(),
		UserID: targetUser.ID,
	}
	c.Api.UnbanChatMember(c.Ctx, unbanParams)

	fullName := targetUser.FirstName
	if targetUser.LastName != "" {
		fullName += " " + targetUser.LastName
	}

	ap.Log.Info().
		Int64("user_id", targetUser.ID).
		Str("username", targetUser.Username).
		Msg("踢出用户")

	c.Replyf("✅ 已将 %s 踢出群组", fullName)
}

// -------------------- 辅助函数 --------------------

// 获取目标用户
func (ap *AdminPlugin) getTargetUser(c *context.Context, msg *telego.Message) *telego.User {
	// 1. 检查是否回复了消息
	if msg.ReplyToMessage != nil {
		return msg.ReplyToMessage.From
	}

	// 2. 检查是否提到了用户
	if len(msg.Entities) > 0 {
		for _, entity := range msg.Entities {
			// TextMention: 点击用户名会跳转到用户资料（有 User 对象）
			if entity.Type == telego.EntityTypeTextMention && entity.User != nil {
				return entity.User
			}

			// Mention: 普通的 @username（需要通过 ChatMember 查询）
			if entity.Type == telego.EntityTypeMention {
				// 提取 username（去掉 @）
				username := msg.Text[entity.Offset+1 : entity.Offset+entity.Length]

				// 注意：Telegram Bot API 不支持直接通过 username 查询成员

				ap.Log.Warn().Str("username", username).Msg("无法通过 @username 直接获取用户信息")
				return nil
			}
		}
	}

	return nil
}
