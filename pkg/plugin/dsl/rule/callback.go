package rule

import "yueling_tg/internal/core/context"

func CallBackStartsWith(prefix ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		return startsWith(ctx.GetCallbackData(), prefix...)
	})
}
func CallbackFullMatch(prefix ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		return fullMatch(ctx.GetCallbackData(), prefix...)
	})
}

func CallbackKeyword(keyword ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		return keywords(ctx.GetCallbackData(), keyword...)
	})
}

func IsInlineQueryEvent() Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		return ctx.GetInlineQuery() != nil
	})
}
