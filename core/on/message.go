package on

import (
	"yueling_tg/core/handler"
	"yueling_tg/core/plugin"
	"yueling_tg/core/provider"
	"yueling_tg/core/rule"
)

// 事件
func OnMessage(handler *handler.Handler) *plugin.Matcher {
	handler.RegisterDynamicProvider(provider.MessageProvider())
	return plugin.NewMatcher(rule.IsMessageEvent(), handler)
}

// 命令
func OnStartsWith(prefixes []string, handler *handler.Handler) *plugin.Matcher {
	return plugin.NewMatcher(rule.StartsWith(prefixes...), handler)
}

func OnEndsWith(suffixes []string, handler *handler.Handler) *plugin.Matcher {
	return plugin.NewMatcher(rule.EndsWith(suffixes...), handler)
}

func OnFullMatch(patterns []string, handler *handler.Handler) *plugin.Matcher {
	return plugin.NewMatcher(rule.FullMatch(patterns...), handler)
}

func OnKeyword(keywords []string, handler *handler.Handler) *plugin.Matcher {
	return plugin.NewMatcher(rule.Keyword(keywords...), handler)
}

func OnCommand(cmds []string, caseSensitive bool, handler *handler.Handler) *plugin.Matcher {
	handler.RegisterDynamicProviders(provider.CommandArgsProvider(cmds), provider.CommandContextProvider(cmds))
	return plugin.NewMatcher(rule.Command(caseSensitive, cmds...), handler)
}

func OnRegex(patterns []string, handler *handler.Handler) *plugin.Matcher {
	return plugin.NewMatcher(rule.Regex(patterns...), handler)
}
