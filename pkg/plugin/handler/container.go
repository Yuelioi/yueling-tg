package handler

import (
	"reflect"
	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin/provider"
)

// 全局容器实例（应用启动时初始化一次）
var GlobalContainer *Container

// Container 依赖注入容器（应用级别，全局单例）
type Container struct {
	// 静态 Provider（启动时注册，整个应用生命周期有效）
	staticProviders []provider.Provider

	// 动态 Provider（每次请求时执行）
	dynamicProviders []provider.Provider
}

// NewContainer 创建新的容器
func NewContainer() *Container {
	return &Container{
		staticProviders:  make([]provider.Provider, 0),
		dynamicProviders: make([]provider.Provider, 0),
	}
}

// RegisterStatic 注册静态依赖（如：数据库连接、配置等）
func (c *Container) RegisterStatic(providers ...provider.Provider) *Container {
	c.staticProviders = append(c.staticProviders, providers...)
	return c
}

// RegisterDynamic 注册动态依赖（如：Context、Request 相关等）
func (c *Container) RegisterDynamic(providers ...provider.Provider) *Container {
	c.dynamicProviders = append(c.dynamicProviders, providers...)
	return c
}

// NewResolver 创建依赖解析器（单容器版本）
func (c *Container) NewResolver(ctx *context.Context) *Resolver {
	return &Resolver{
		container: c,
		ctx:       ctx,
		cache:     make(map[reflect.Type]reflect.Value),
	}
}

// NewMergedResolver 创建合并多个容器的解析器
// 优先级：当前容器 > 父容器（从左到右）
func (c *Container) NewMergedResolver(ctx *context.Context, parentContainers ...*Container) *Resolver {
	return &Resolver{
		container:        c,
		parentContainers: parentContainers,
		ctx:              ctx,
		cache:            make(map[reflect.Type]reflect.Value),
	}
}

// InitGlobalContainer 初始化全局容器（应用启动时调用一次）
func InitGlobalContainer() *Container {
	GlobalContainer = NewContainer()

	return GlobalContainer
}
