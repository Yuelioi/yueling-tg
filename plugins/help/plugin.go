package help

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"yueling_tg/core/bot"
	"yueling_tg/core/context"
	"yueling_tg/core/handler"
	"yueling_tg/core/on"
	"yueling_tg/core/params"
	"yueling_tg/core/plugin"
)

var _ plugin.Plugin = (*helper)(nil)

type helper struct {
	*plugin.Base
}

func New() plugin.Plugin {

	base := plugin.NewBase(&plugin.PluginInfo{
		ID:          "help",
		Name:        "帮助插件",
		Description: "提供帮助信息",
		Version:     "0.1.0",
		Author:      "月离",
		Usage:       "help [插件ID]",
		Group:       "builtin",
		Extra:       nil,
	})

	h := &helper{base}

	helpHandler := handler.NewHandler(h.listPlugins)
	helpMatcher := on.OnCommand([]string{"help", "帮助"}, true, helpHandler)
	h.SetMatchers([]*plugin.Matcher{helpMatcher})

	return h
}

func (h *helper) listPlugins(ctx *context.Context, cmdCtx params.CommandContext) {
	plugins := bot.Plugins()
	if plugins == nil {
		h.Log.Warn().Msg("Plugins() 返回 nil")
		ctx.Send("❌ 当前没有可用的插件")
		return
	}

	// 按插件名排序，保证顺序固定
	var sortedPlugins []plugin.Plugin
	for _, p := range plugins {
		if p != nil && p.PluginInfo() != nil {
			sortedPlugins = append(sortedPlugins, p)
		}
	}
	sort.Slice(sortedPlugins, func(i, j int) bool {
		return sortedPlugins[i].PluginInfo().Name < sortedPlugins[j].PluginInfo().Name
	})

	// 处理命令参数
	if cmdCtx.Args.Len() != 0 {
		arg := cmdCtx.Args.Get(0)
		// 尝试将参数解析为数字 ID
		if id, err := strconv.Atoi(arg); err == nil {
			if id >= 1 && id <= len(sortedPlugins) {
				target := sortedPlugins[id-1]
				info := target.PluginInfo()
				ctx.Send(fmt.Sprintf(
					"📖 插件 #%d '%s'\n描述: %s\n用法: %s",
					id, info.Name, info.Description, info.Usage,
				))
				return
			} else {
				ctx.Send(fmt.Sprintf("❌ 插件 ID '%d' 不存在", id))
				return
			}
		}

		// 非数字 → 按名称查找
		var target plugin.Plugin
		for _, p := range sortedPlugins {
			if p.PluginInfo().Name == arg {
				target = p
				break
			}
		}
		if target != nil {
			info := target.PluginInfo()
			ctx.Send(fmt.Sprintf(
				"📖 插件 '%s'\n描述: %s\n用法: %s",
				info.Name, info.Description, info.Usage,
			))
		} else {
			ctx.Sendf("❌ 未找到名为『%s』的插件", arg)
		}
		return
	}

	// 没有参数 → 列出插件列表并显示 ID
	var msgs strings.Builder
	msgs.WriteString("✨ 可用插件列表:\n")
	msgs.WriteString("使用help <插件ID> 获取插件详细信息\n")
	for i, p := range sortedPlugins {
		info := p.PluginInfo()
		name := "<未知>"
		if info != nil && info.Name != "" {
			name = info.Name
		}

		msgs.WriteString(fmt.Sprintf("🔹 #%d %s \n", i+1, name))
	}

	ctx.Send(msgs.String())
}
