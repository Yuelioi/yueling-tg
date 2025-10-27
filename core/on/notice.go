package on

import (
	"yueling_tg/core/handler"
	"yueling_tg/core/plugin"
	"yueling_tg/core/rule"
)

func OnNotice(handler *handler.Handler) *plugin.Matcher {
	return plugin.NewMatcher(rule.IsNoticeEvent(), handler)
}
