package condition

import (
	"yueling_tg/internal/core/context"
)

type multiCondition struct {
	conditions []Condition
	combiner   func(results []bool) bool
}

func (mc *multiCondition) Match(ctx *context.Context) bool {
	results := make([]bool, len(mc.conditions))
	for i, cond := range mc.conditions {
		results[i] = cond.Match(ctx)
	}
	return mc.combiner(results)
}

type singleCondition struct {
	condition Condition
	modifier  func(result bool) bool
}

func (sc *singleCondition) Match(ctx *context.Context) bool {
	originalResult := sc.condition.Match(ctx)
	return sc.modifier(originalResult)
}
