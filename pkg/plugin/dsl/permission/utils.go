package permission

import (
	"log"
	"yueling_tg/internal/core/context"

	"github.com/mymmrac/telego"
)

func getUserRole(ctx *context.Context) string {
	// 1. 确保有消息和发送者
	if ctx.GetMessage() == nil || ctx.GetUser() == nil {
		return "unknown"
	}

	// 2. 检查聊天类型：私聊中用户始终是 'member' 或 'owner'
	chatType := ctx.GetChatType()
	if chatType == "private" {
		// 在私聊中，用户就是聊天对象，可以视为 'owner' 或 'self'
		return "self"
	}

	// 如果不是群组/频道，也返回默认角色
	if chatType != "group" && chatType != "supergroup" && chatType != "channel" {
		return "member"
	}

	// 3. 调用 API 获取成员信息
	member, err := ctx.Api.GetChatMember(ctx.Ctx, &telego.GetChatMemberParams{
		ChatID: ctx.GetChatID(), // 聊天的ID
		UserID: ctx.GetUserID(), // 用户的ID
	})
	if err != nil {
		log.Printf("调用 GetChatMember 失败: %v", err)
		return "unknown"
	}

	// 4. 解析成员状态
	// member 是一个 ChatMember 接口，需要通过类型断言获取具体类型
	var role string
	switch m := member.(type) {
	case *telego.ChatMemberOwner:
		role = "creator"
	case *telego.ChatMemberAdministrator:
		role = "admin"
	case *telego.ChatMemberMember:
		role = "member"
	case *telego.ChatMemberRestricted:
		role = "restricted"
	case *telego.ChatMemberLeft:
		role = "left"
	case *telego.ChatMemberBanned:
		role = "kicked"
	default:
		log.Printf("未知的 ChatMember 类型: %T", m)
		role = "unknown"
	}

	// 5. 映射角色到统一格式
	switch role {
	case "creator":
		return "creator" // 聊天创建者
	case "admin":
		return "admin" // 管理员
	case "member":
		return "member" // 普通成员
	case "restricted":
		return "restricted" // 受限成员（被禁言等）
	case "left", "kicked":
		return "not_in_chat" // 已经不在群里
	default:
		return "member"
	}
}
