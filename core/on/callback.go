package on

import (
	"yueling_tg/core/handler"
	"yueling_tg/core/plugin"
	"yueling_tg/core/provider"
	"yueling_tg/core/rule"
)

func OnCallback(handler *handler.Handler) *plugin.Matcher {
	handler.RegisterDynamicProvider(provider.CallbackDataProvider())
	return plugin.NewMatcher(rule.IsCallbackEvent(), handler)
}

func OnCallbackFullMatch(patterns []string, handler *handler.Handler) *plugin.Matcher {
	handler.RegisterDynamicProvider(provider.CallbackDataProvider())
	return plugin.NewMatcher(rule.CallbackFullMatch(patterns...), handler)
}

func OnCallbackStartsWith(patterns []string, handler *handler.Handler) *plugin.Matcher {
	handler.RegisterDynamicProvider(provider.CallbackDataProvider())
	return plugin.NewMatcher(rule.CallBackStartsWith(patterns...), handler)
}
