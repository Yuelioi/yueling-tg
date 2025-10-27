package permission

import (
	"log"
	"strings"
	"yueling_tg/core/context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	// 3. 准备 GetChatMember 配置
	config := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: ctx.GetChatID(), // 聊天的ID
			UserID: ctx.GetUserID(), // 用户的ID
		},
	}

	// 4. 调用 API 获取成员信息
	member, err := ctx.Api.GetChatMember(config)
	if err != nil {
		log.Printf("调用 GetChatMember 失败: %v", err)
		return "unknown"
	}

	// 5. 解析 Status 字段
	// Status 字段返回的是一个字符串，可能的取值包括:
	// "creator", "administrator", "member", "restricted", "left", "kicked"
	role := strings.ToLower(member.Status)

	// 为了简化，你可以将不同的状态映射到你需要的角色：
	switch role {
	case "creator":
		return "creator" // 聊天创建者
	case "administrator":
		return "admin" // 管理员
	case "member":
		return "member" // 普通成员
	case "restricted":
		return "restricted" // 受限成员（被禁言等）
	case "left", "kicked":
		return "not_in_chat" // 已经不在群里
	case "unknown":
		return "unknown" // 报错
	default:
		return "member"
	}
}
