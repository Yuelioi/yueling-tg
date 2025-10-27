package on

import (
	"yueling_tg/core/handler"
	"yueling_tg/core/plugin"
	"yueling_tg/core/rule"
)

func On(rule rule.Rule, handler *handler.Handler) *plugin.Matcher {
	return plugin.NewMatcher(rule, handler)
}
