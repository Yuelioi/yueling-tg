package ban

import (
	"math/rand"
	"time"

	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/handler"
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
	username := ctx.GetFullName()

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

	if ctx.IsPrivate() || ctx.IsAdmin() {
		ctx.Sendf("%s %s，%d小时后见！😴💤", username, sleepWord, sleepHours)
		return
	}

	if !ctx.IsBotAdmin() || !ctx.CanBotRestrictMembers() {
		p.Log.Error().Msg("获取机器人权限失败")
		ctx.Reply("我没有管理员权限，无法让你睡觉哦~")
		return
	}

	// 禁言用户
	ctx.MuteUser(ctx.GetUserID(), time.Duration(sleepSeconds)*time.Second)

	p.Log.Info().
		Int64("user_id", userID).
		Str("username", username).
		Int("sleep_hours", sleepHours).
		Msg("禁言成功")

	ctx.Replyf("%s %s，%d小时后见！😴💤", username, sleepWord, sleepHours)

}
