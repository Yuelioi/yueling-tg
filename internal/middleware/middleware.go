package middleware

import (
	"yueling_tg/internal/core/context"
)

// Middleware 中间件接口
type Middleware interface {
	// Process 中间件处理函数
	Process(ctx *context.Context, next HandlerFunc) error

	// Name 中间件名称
	Name() string
}

type HandlerFunc func(ctx *context.Context) error
