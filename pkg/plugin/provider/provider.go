package provider

import (
	"yueling_tg/internal/core/context"
)

// Provider 接口用于提供依赖
type Provider interface {
	Provide(ctx *context.Context) any
}

// 函数式 Provider
var _ Provider = StaticProvider(nil)
var _ Provider = DynamicProvider(nil)

// 静态 Provider
type StaticProvider func() any

func (f StaticProvider) Provide(ctx *context.Context) any {
	return f()
}

// 动态 Provider
type DynamicProvider func(ctx *context.Context) any

func (f DynamicProvider) Provide(ctx *context.Context) any {
	return f(ctx)
}
