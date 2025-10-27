package rule

import (
	"yueling_tg/core/context"

	"yueling_tg/core/condition"
)

// 规则
type Rule condition.Condition

// RuleFunc 规则函数类型
type RuleFunc func(ctx *context.Context) bool

func (f RuleFunc) Match(ctx *context.Context) bool {
	return f(ctx)
}
