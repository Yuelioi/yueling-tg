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
	"yueling_tg/core/context"
	"yueling_tg/core/provider"
)

// Handler 事件处理函数
type Handler struct {
	fn         any
	fnValue    reflect.Value
	fnType     reflect.Type
	paramTypes []reflect.Type
	container  *Container
}

// NewHandler 创建处理器
func NewHandler(fn any) *Handler {

	container := NewContainer()

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

// Call 执行处理器
func (h *Handler) Call(ctx *context.Context, providers ...provider.Provider) error {
	// 创建请求级别的解析器
	h.container.RegisterDynamic(provider.ContextProvider(ctx))

	resolver := h.container.NewResolver(ctx)

	// 注入依赖
	for _, p := range providers {
		resolver.container.RegisterStatic(p)
	}

	// 注入全局依赖
	for _, p := range h.container.dynamicProviders {
		resolver.container.RegisterStatic(p)
	}

	// 解析所有参数
	args, err := resolver.ResolveAll(h.paramTypes)

	if err != nil {
		return fmt.Errorf("解析参数失败: %w", err)
	}

	// args = append(args, reflect.ValueOf(ctx))

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

func (h *Handler) RegisterDynamicProvider(provider provider.Provider) *Handler {
	h.container.RegisterDynamic(provider)
	return h
}
func (h *Handler) RegisterDynamicProviders(providers ...provider.Provider) *Handler {
	for _, provider := range providers {
		h.container.RegisterDynamic(provider)
	}
	return h
}
