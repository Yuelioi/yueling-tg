package middleware

import (
	"yueling_tg/core/context"
)

// 构建中间件 并返回最终处理函数
func Chain(middlewares []Middleware, final func(ctx *context.Context) error) func(ctx *context.Context) error {
	if len(middlewares) == 0 {
		return final
	}
	current := final
	for i := len(middlewares) - 1; i >= 0; i-- {
		m := middlewares[i]
		next := current
		current = func(ctx *context.Context) error {
			return m.Process(ctx, next)
		}
	}
	return current
}
