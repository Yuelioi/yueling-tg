package rule

import (
	"yueling_tg/internal/core/context"
)

func IsCommandEvent() RuleFunc {
	return func(ctx *context.Context) bool {
		return ctx.IsCommand()
	}
}

func IsMessageEvent() RuleFunc {
	return func(ctx *context.Context) bool {
		return ctx.IsMessage()
	}
}

func IsNoticeEvent() RuleFunc {
	return func(ctx *context.Context) bool {
		return ctx.IsNotice()
	}
}

func IsCallbackEvent() RuleFunc {
	return func(ctx *context.Context) bool {
		return ctx.IsCallback()
	}
}
