package handler

import (
	"reflect"
	"yueling_tg/core/context"
	"yueling_tg/core/provider"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// 依赖注入容器（应用级别，全局单例）
type Container struct {
	// 静态 Provider（启动时注册，整个应用生命周期有效）
	staticProviders []provider.Provider

	// 动态 Provider（每次请求时执行）
	dynamicProviders []provider.Provider
}

// 创建新的容器
func NewContainer() *Container {
	return &Container{
		staticProviders:  make([]provider.Provider, 0),
		dynamicProviders: make([]provider.Provider, 0),
	}
}

// 注册静态依赖（如：数据库连接、配置等）
func (c *Container) RegisterStatic(providers ...provider.Provider) *Container {
	c.staticProviders = append(c.staticProviders, providers...)
	return c
}

// 注册动态依赖（如：Context、Request 相关等）
func (c *Container) RegisterDynamic(providers ...provider.Provider) *Container {
	c.dynamicProviders = append(c.dynamicProviders, providers...)
	return c
}

// 创建依赖解析器
func (c *Container) NewResolver(ctx *context.Context) *Resolver {
	return &Resolver{
		container: c,
		ctx:       ctx,
		cache:     make(map[reflect.Type]reflect.Value),
	}
}

// InitContainer 初始化全局容器
func InitContainer(api *tgbotapi.BotAPI) *Container {
	globalContainer := NewContainer()

	// 注册静态依赖（整个应用生命周期）
	globalContainer.RegisterStatic(
		provider.StaticProvider(func(ctx *context.Context) any {
			return api
		}),
		// 其他静态依赖...
	)

	// 注册动态依赖（每次请求时执行）
	globalContainer.RegisterDynamic(
		// Context 本身
		provider.DynamicProvider(func(ctx *context.Context) any {
			return ctx
		}),
		// 命令参数
		provider.DynamicProvider(func(ctx *context.Context) any {
			if ctx.IsCommand() {
				return ctx.GetCommandArgs()
			}
			return ""
		}),
		// 其他动态依赖...
	)

	return globalContainer
}
