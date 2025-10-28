package ban

import (
	"math/rand"
	"time"

	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/handler"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var _ plugin.Plugin = (*SleepPlugin)(nil)

// -------------------- 插件结构 --------------------

type SleepPlugin struct {
	*plugin.Base
	sleepWords []string
}

var pluginInfo = &plugin.PluginInfo{
	ID:          "sleep",
	Name:        "我要睡觉",
	Description: "别水群了，赶紧睡觉，强制睡眠",
	Version:     "1.0.0",
	Author:      "月离",
	Usage:       "我要睡觉",
	Group:       "娱乐",
	Extra:       make(map[string]any),
}

func New() plugin.Plugin {
	p := &SleepPlugin{
		Base: plugin.NewBase(pluginInfo),
		sleepWords: []string{
			"被梦魇抓走了",
			"被僵尸吃掉了脑子",
			"被外星人抓走做实验了",
			"去梦里拯救世界了",
			"被催眠了",
			"睡着了",
			"进入梦境了",
			"被睡神带走了",
		},
	}

	sleepHandler := handler.NewHandler(p.handleSleep)
	sleepMatcher := plugin.OnFullMatch([]string{"我要睡觉"}, sleepHandler).
		SetPriority(10)

	p.AddMatcher(sleepMatcher)

	return p
}

// -------------------- 处理器 --------------------

func (p *SleepPlugin) handleSleep(ctx *context.Context) {

	// 获取用户信息
	userID := ctx.GetUserID()
	username := ctx.GetNickName()
	chatId := ctx.GetChatID()

	// 随机睡眠时间（5-8小时，转换为秒）
	sleepHours := rand.Intn(4) + 5 // 5-8小时
	sleepSeconds := sleepHours * 60 * 60

	// 随机选择睡眠理由
	sleepWord := p.sleepWords[rand.Intn(len(p.sleepWords))]

	p.Log.Info().
		Int64("user_id", userID).
		Str("username", username).
		Int("sleep_hours", sleepHours).
		Msg("用户要睡觉")

	// 检查是否是群组
	if ctx.IsGroup() || ctx.IsSuperGroup() {
		// 检查机器人是否是管理员
		botMember, err := ctx.Api.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: chatId,
				UserID: userID,
			},
		})

		if err != nil {
			p.Log.Error().Err(err).Msg("获取机器人权限失败")
			ctx.Reply("我没有管理员权限，无法让你睡觉哦~")
			return
		}

		// 检查机器人是否有禁言权限
		if !botMember.CanRestrictMembers {
			ctx.Reply("我没有禁言权限，无法让你睡觉哦~")
			return
		}

		// 检查用户是否是管理员
		userMember, err := ctx.Api.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
				ChatID: chatId,
				UserID: userID,
			},
		})

		if err != nil {
			p.Log.Error().Err(err).Msg("获取用户权限失败")
			ctx.Reply("获取你的权限信息失败了~")
			return
		}

		// 如果用户是管理员或创建者，不禁言
		if userMember.IsAdministrator() || userMember.IsCreator() {
			ctx.Replyf("%s 是管理员，不能被强制睡觉哦~ 😴", username)
			return
		}

		// 禁言用户
		until := time.Now().Add(time.Duration(sleepSeconds) * time.Second)
		restrictConfig := tgbotapi.RestrictChatMemberConfig{
			ChatMemberConfig: tgbotapi.ChatMemberConfig{
				ChatID: chatId,
				UserID: userID,
			},
			UntilDate: until.Unix(),
			Permissions: &tgbotapi.ChatPermissions{
				CanSendMessages:       false,
				CanSendMediaMessages:  false,
				CanSendPolls:          false,
				CanSendOtherMessages:  false,
				CanAddWebPagePreviews: false,
				CanChangeInfo:         false,
				CanInviteUsers:        false,
				CanPinMessages:        false,
			},
		}

		_, err = ctx.Api.Request(restrictConfig)
		if err != nil {
			p.Log.Error().Err(err).Msg("禁言失败")
			ctx.Reply("禁言失败了，可能是权限不足~")
			return
		}

		p.Log.Info().
			Int64("user_id", userID).
			Str("username", username).
			Int("sleep_hours", sleepHours).
			Msg("禁言成功")

		ctx.Replyf("%s %s，%d小时后见！😴💤", username, sleepWord, sleepHours)
	} else {
		// 私聊情况，只是发送消息
		ctx.Replyf("你%s，%d小时后见！😴💤\n（私聊我无法让你真的睡觉哦~）", sleepWord, sleepHours)
	}
}
