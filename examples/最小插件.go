package examples

import (
	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/dsl/rule"
	"yueling_tg/pkg/plugin/handler"
)

var _ plugin.Plugin = (*PluginExample3)(nil)

type PluginExample3 struct {
}

func Print(ctx *context.Context) {
	ctx.Reply("pong")
}

func (p *PluginExample3) Matchers() []*plugin.Matcher {

	h := handler.NewHandler(Print)
	m := plugin.NewMatcher(rule.StartsWith("ping"), h)
	return []*plugin.Matcher{m}

}

// 插件信息
func (p *PluginExample3) PluginInfo() *plugin.PluginInfo {
	return &plugin.PluginInfo{
		ID:          "",
		Name:        "",
		Description: "",
		Version:     "",
		Author:      "",
		Usage:       "",
		Examples:    []string{},
		Group:       "",
		Extra:       map[string]any{},
	}
}
