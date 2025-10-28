package rule

import (
	"yueling_tg/internal/core/context"

	"yueling_tg/pkg/plugin/dsl/condition"
)

// 规则
type Rule = condition.Condition

// RuleFunc 规则函数类型
type RuleFunc func(ctx *context.Context) bool

func (f RuleFunc) Match(ctx *context.Context) bool {
	return f(ctx)
}
