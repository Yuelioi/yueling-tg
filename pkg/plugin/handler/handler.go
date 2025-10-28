// Package handler 封装了事件处理函数的逻辑，支持依赖注入与自动调用。
//
// 核心功能：
//   - 使用 NewHandler 构造并注册事件处理器
//   - 使用 RegisterProvider 方法注入函数依赖
//   - 使用 Call 方法执行并注入上下文和依赖参数
//
// 本包适用于构建具有自动依赖注入能力的事件驱动系统。

package handler

import (
	"fmt"
	"reflect"
	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin/provider"
)

// Handler 事件处理函数
type Handler struct {
	fn         any
	fnValue    reflect.Value
	fnType     reflect.Type
	paramTypes []reflect.Type
	container  *Container // 插件级别的容器
}

// NewHandler 创建处理器（带插件级容器）
func NewHandler(fn any) *Handler {
	container := NewContainer() // 每个 Handler 独立的容器

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		panic("NewHandler: 参数必须是函数")
	}

	numParams := fnType.NumIn()
	paramTypes := make([]reflect.Type, numParams)
	for i := 0; i < numParams; i++ {
		paramTypes[i] = fnType.In(i)
	}

	return &Handler{
		fn:         fn,
		fnValue:    fnValue,
		fnType:     fnType,
		paramTypes: paramTypes,
		container:  container,
	}
}

// Call 执行处理器（合并全局容器和插件容器）
func (h *Handler) Call(ctx *context.Context, providers ...provider.Provider) error {
	var resolver *Resolver

	// 如果有临时 providers，创建临时容器并设置为最高优先级
	if len(providers) > 0 {
		tempContainer := NewContainer()
		tempContainer.RegisterStatic(providers...)

		// 创建解析器：临时容器 > 插件容器 > 全局容器
		resolver = tempContainer.NewMergedResolver(ctx, h.container, GlobalContainer)
	} else {
		// 没有临时 providers，直接合并：插件容器 > 全局容器
		resolver = h.container.NewMergedResolver(ctx, GlobalContainer)
	}

	// 解析所有参数
	args, err := resolver.ResolveAll(h.paramTypes)
	if err != nil {
		return fmt.Errorf("解析参数失败: %w", err)
	}

	// 调用函数
	results := h.fnValue.Call(args)

	// 处理返回值
	if len(results) > 0 {
		if last := results[len(results)-1]; !last.IsNil() {
			if err, ok := last.Interface().(error); ok {
				return err
			}
		}
	}

	return nil
}

// RegisterDynamicProvider 向插件容器注册动态依赖
func (h *Handler) RegisterDynamicProvider(p provider.Provider) *Handler {
	h.container.RegisterDynamic(p)
	return h
}

// RegisterDynamicProviders 批量注册动态依赖
func (h *Handler) RegisterDynamicProviders(providers ...provider.Provider) *Handler {
	h.container.RegisterDynamic(providers...)
	return h
}

// RegisterStaticProvider 向插件容器注册静态依赖
func (h *Handler) RegisterStaticProvider(p provider.Provider) *Handler {
	h.container.RegisterStatic(p)
	return h
}

// RegisterStaticProviders 批量注册静态依赖
func (h *Handler) RegisterStaticProviders(providers ...provider.Provider) *Handler {
	h.container.RegisterStatic(providers...)
	return h
}
