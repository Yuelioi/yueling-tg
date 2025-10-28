package recall

import (
	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/handler"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var _ plugin.Plugin = (*RecallPlugin)(nil)

// -------------------- 插件结构 --------------------

type RecallPlugin struct {
	*plugin.Base
}

func New() plugin.Plugin {
	rp := &RecallPlugin{
		Base: plugin.NewBase(&plugin.PluginInfo{
			ID:          "recall",
			Name:        "撤回消息",
			Description: "撤回你发送的消息，并回复提示",
			Version:     "1.0.0",
			Author:      "月离",
			Usage:       "撤回",
			Group:       "娱乐",
			Extra:       make(map[string]any),
		},
		),
	}

	recallHandler := handler.NewHandler(rp.handleRecall)
	recallMatcher := plugin.OnFullMatch([]string{"撤回"}, recallHandler).
		SetPriority(10).SetBlock(true)

	rp.AddMatcher(recallMatcher)
	return rp

}

// -------------------- 处理器 --------------------
func (rp *RecallPlugin) handleRecall(ctx *context.Context) {
	msg := ctx.GetMessage()
	if msg == nil {
		return
	}

	username := ctx.GetNickName()

	// 确认用户是回复了一条消息
	if msg.ReplyToMessage == nil {
		ctx.Reply("⚠ 请回复你想撤回的消息，然后发送“撤回”命令")
		return
	}

	targetMsg := msg.ReplyToMessage // 用户想撤回的消息

	// 群组中检查机器人权限
	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		botMember, err := ctx.Api.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: msg.Chat.ID,
				UserID: ctx.Api.Self.ID,
			},
		})
		if err != nil {
			rp.Log.Error().Err(err).Msg("获取机器人权限失败")
			ctx.Reply("❌ 无法获取我的管理员权限，撤回失败~")
			return
		}

		if !botMember.CanDeleteMessages {
			ctx.Reply("⚠ 我没有删除消息权限，无法撤回~")
			return
		}

		// 撤回目标消息
		deleteMsg := tgbotapi.DeleteMessageConfig{
			ChatID:    msg.Chat.ID,
			MessageID: targetMsg.MessageID,
		}
		_, err = ctx.Api.Request(deleteMsg)
		if err != nil {
			rp.Log.Error().Err(err).Msg("撤回消息失败")
			ctx.Reply("❌ 撤回失败了，可能权限不足~")
			return
		}

		rp.Log.Info().
			Str("username", username).
			Int("message_id", targetMsg.MessageID).
			Msg("消息撤回成功")

	} else {
		// 私聊无法撤回，只发送提示
		ctx.Reply("⚠ 私聊无法撤回消息，但我看到了你的请求~")
	}
}
