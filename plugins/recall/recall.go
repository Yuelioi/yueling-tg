package recall

import (
	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"
)

var _ plugin.Plugin = (*RecallPlugin)(nil)

// -------------------- 插件结构 --------------------

type RecallPlugin struct {
	*plugin.Base
}

func New() plugin.Plugin {
	rp := &RecallPlugin{}

	info := &plugin.PluginInfo{
		ID:          "recall",
		Name:        "撤回消息",
		Description: "撤回你发送的消息，并回复提示",
		Version:     "1.0.0",
		Author:      "月离",
		Usage:       "撤回",
		Group:       "娱乐",
		Extra:       make(map[string]any),
	}

	builder := plugin.New().Info(info)

	builder.OnFullMatch("撤回").Block(true).Do(rp.handleRecall)

	return builder.Go(rp)
}

// -------------------- 处理器 --------------------
func (rp *RecallPlugin) handleRecall(ctx *context.Context) {
	msg := ctx.GetMessage()
	if msg == nil {
		return
	}

	username := ctx.GetFullName()

	// 确认用户是回复了一条消息
	if msg.ReplyToMessage == nil {
		return
	}

	targetMsg := msg.ReplyToMessage

	// 群组中检查机器人权限
	if ctx.IsGroup() || ctx.IsSuperGroup() {
		if !ctx.CanBotDeleteMessage() {
			rp.Log.Error().Msg("获取机器人权限失败")
			ctx.Reply("❌ 尚未取得管理员权限，撤回失败~")
			return
		}

		err := ctx.DeleteMessage(targetMsg.MessageID)
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
		ctx.Reply("⚠ 私聊无法撤回消息~")
	}
}
