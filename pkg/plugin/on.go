package plugin

import (
	"yueling_tg/pkg/plugin/dsl/rule"
	"yueling_tg/pkg/plugin/handler"
	"yueling_tg/pkg/plugin/provider"
)

func On(rule rule.Rule, handler *handler.Handler) *Matcher {
	return NewMatcher(rule, handler)
}

func OnCallback(handler *handler.Handler) *Matcher {
	handler.RegisterDynamicProvider(provider.CallbackDataProvider())
	return NewMatcher(rule.IsCallbackEvent(), handler)
}

func OnCallbackFullMatch(patterns []string, handler *handler.Handler) *Matcher {
	handler.RegisterDynamicProvider(provider.CallbackDataProvider())
	return NewMatcher(rule.CallbackFullMatch(patterns...), handler)
}

func OnCallbackStartsWith(patterns []string, handler *handler.Handler) *Matcher {
	handler.RegisterDynamicProvider(provider.CallbackDataProvider())
	return NewMatcher(rule.CallBackStartsWith(patterns...), handler)
}

func OnNotice(handler *handler.Handler) *Matcher {
	return NewMatcher(rule.IsNoticeEvent(), handler)
}

func OnMessage(handler *handler.Handler) *Matcher {
	handler.RegisterDynamicProvider(provider.MessageProvider())
	return NewMatcher(rule.IsMessageEvent(), handler)
}

// 命令
func OnStartsWith(prefixes []string, handler *handler.Handler) *Matcher {
	return NewMatcher(rule.StartsWith(prefixes...), handler)
}

func OnEndsWith(suffixes []string, handler *handler.Handler) *Matcher {
	return NewMatcher(rule.EndsWith(suffixes...), handler)
}

func OnFullMatch(patterns []string, handler *handler.Handler) *Matcher {
	return NewMatcher(rule.FullMatch(patterns...), handler)
}

func OnKeyword(keywords []string, handler *handler.Handler) *Matcher {
	return NewMatcher(rule.Keyword(keywords...), handler)
}

func OnCommand(cmds []string, caseSensitive bool, handler *handler.Handler) *Matcher {
	handler.RegisterDynamicProviders(provider.CommandArgsProvider(cmds), provider.CommandContextProvider(cmds))
	return NewMatcher(rule.Command(caseSensitive, cmds...), handler)
}

func OnRegex(patterns []string, handler *handler.Handler) *Matcher {
	return NewMatcher(rule.Regex(patterns...), handler)
}

// OnInlineQuery 创建一个 InlineQuery Matcher
func OnInlineQuery(handler *handler.Handler) *Matcher {
	handler.RegisterDynamicProvider(provider.InlineQueryProvider())
	return NewMatcher(rule.IsInlineQueryEvent(), handler)
}
