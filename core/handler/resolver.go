package handler

import (
	"fmt"
	"reflect"
	"yueling_tg/core/context"
)

// 依赖解析器（请求级别，每次请求创建新实例）
type Resolver struct {
	container *Container
	ctx       *context.Context
	cache     map[reflect.Type]reflect.Value
}

// 解析指定类型的依赖
func (r *Resolver) Resolve(t reflect.Type) (reflect.Value, error) {
	// 1. 检查缓存
	if v, ok := r.cache[t]; ok {
		return v, nil
	}

	// 2. 尝试从静态 Provider 解析
	for _, p := range r.container.staticProviders {
		v := p.Provide(r.ctx)
		if v == nil {
			continue
		}

		vt := reflect.TypeOf(v)
		vv := reflect.ValueOf(v)

		// 尝试类型匹配（带自动转换）
		convertedValue, ok := tryConvertType(vv, vt, t)
		if ok {
			r.cache[t] = convertedValue
			return convertedValue, nil
		}
	}

	// 3. 尝试从动态 Provider 解析
	for _, p := range r.container.dynamicProviders {
		v := p.Provide(r.ctx)
		if v == nil {
			continue
		}

		vt := reflect.TypeOf(v)
		vv := reflect.ValueOf(v)

		// 尝试类型匹配（带自动转换）
		convertedValue, ok := tryConvertType(vv, vt, t)
		if ok {
			r.cache[t] = convertedValue
			return convertedValue, nil
		}
	}

	return reflect.Value{}, fmt.Errorf("无法解析类型: %v", t)
}

// ResolveAll 解析多个类型
func (r *Resolver) ResolveAll(types []reflect.Type) ([]reflect.Value, error) {
	values := make([]reflect.Value, len(types))
	for i, t := range types {
		v, err := r.Resolve(t)
		if err != nil {
			return nil, err
		}
		values[i] = v
	}
	return values, nil
}

// tryConvertType 尝试将源值转换为目标类型（自动处理指针转换）
func tryConvertType(srcValue reflect.Value, srcType, tgtType reflect.Type) (reflect.Value, bool) {
	if srcType == nil || tgtType == nil {
		return reflect.Value{}, false
	}

	// 1. 直接匹配或可赋值
	if srcType == tgtType || srcType.AssignableTo(tgtType) {
		return srcValue, true
	}

	// 2. 接口匹配
	if tgtType.Kind() == reflect.Interface && srcType.Implements(tgtType) {
		return srcValue, true
	}

	// 3. 底层类型相同的命名类型转换（如 type Message tgbotapi.Message）
	if canConvertUnderlyingType(srcValue, srcType, tgtType) {
		return srcValue.Convert(tgtType), true
	}

	// 3. 底层类型相同的命名类型转换（如 type Message tgbotapi.Message）
	if canConvertUnderlyingType(srcValue, srcType, tgtType) {
		return srcValue.Convert(tgtType), true
	}

	// 4. 源是指针，目标不是指针：自动解引用
	if srcType.Kind() == reflect.Ptr && tgtType.Kind() != reflect.Ptr {
		elemType := srcType.Elem()

		// 检查解引用后是否匹配
		if elemType == tgtType || elemType.AssignableTo(tgtType) {
			if !srcValue.IsNil() {
				return srcValue.Elem(), true
			}
		}

		// 检查解引用后底层类型是否相同
		if !srcValue.IsNil() && canConvertUnderlyingType(srcValue.Elem(), elemType, tgtType) {
			return srcValue.Elem().Convert(tgtType), true
		}

		// 检查解引用后是否实现接口
		if tgtType.Kind() == reflect.Interface && elemType.Implements(tgtType) {
			if !srcValue.IsNil() {
				return srcValue.Elem(), true
			}
		}
	}

	// 5. 源不是指针，目标是指针：自动取地址
	if srcType.Kind() != reflect.Ptr && tgtType.Kind() == reflect.Ptr {
		elemType := tgtType.Elem()

		// 检查取地址后是否匹配
		if srcType == elemType || srcType.AssignableTo(elemType) {
			if srcValue.CanAddr() {
				return srcValue.Addr(), true
			}
			// 如果不能直接取地址，创建一个新的可寻址的值
			newValue := reflect.New(srcType)
			newValue.Elem().Set(srcValue)
			return newValue, true
		}

		// 检查底层类型是否相同，需要转换后再取地址
		if canConvertUnderlyingType(srcValue, srcType, elemType) {
			converted := srcValue.Convert(elemType)
			if converted.CanAddr() {
				return converted.Addr(), true
			}
			// 创建新的可寻址值
			newValue := reflect.New(elemType)
			newValue.Elem().Set(converted)
			return newValue, true
		}

		// 检查源类型是否实现目标接口
		if elemType.Kind() == reflect.Interface && srcType.Implements(elemType) {
			if srcValue.CanAddr() {
				return srcValue.Addr(), true
			}
			// 创建新的可寻址值
			newValue := reflect.New(srcType)
			newValue.Elem().Set(srcValue)
			return newValue, true
		}
	}

	// 6. 双方都是指针，但指向不同类型：检查元素类型
	if srcType.Kind() == reflect.Ptr && tgtType.Kind() == reflect.Ptr {
		srcElem := srcType.Elem()
		tgtElem := tgtType.Elem()

		if srcElem == tgtElem || srcElem.AssignableTo(tgtElem) {
			return srcValue, true
		}

		// 底层类型相同的指针转换
		if !srcValue.IsNil() && canConvertUnderlyingType(srcValue.Elem(), srcElem, tgtElem) {
			converted := srcValue.Elem().Convert(tgtElem)
			newPtr := reflect.New(tgtElem)
			newPtr.Elem().Set(converted)
			return newPtr, true
		}

		// 接口检查
		if tgtElem.Kind() == reflect.Interface && srcElem.Implements(tgtElem) {
			return srcValue, true
		}
	}

	return reflect.Value{}, false
}

// canConvertUnderlyingType 检查两个类型的底层类型是否相同且可转换
// 用于处理 type Message tgbotapi.Message 这种命名类型
func canConvertUnderlyingType(srcValue reflect.Value, srcType, tgtType reflect.Type) bool {
	// 必须是同一种 Kind
	if srcType.Kind() != tgtType.Kind() {
		return false
	}

	// 检查是否可以转换
	if !srcValue.CanConvert(tgtType) {
		return false
	}

	// 对于结构体，检查底层类型是否相同
	if srcType.Kind() == reflect.Struct && tgtType.Kind() == reflect.Struct {
		// 如果字段数量、名称、类型都相同，认为是相同的底层类型
		if srcType.NumField() != tgtType.NumField() {
			return false
		}

		for i := 0; i < srcType.NumField(); i++ {
			srcField := srcType.Field(i)
			tgtField := tgtType.Field(i)

			if srcField.Name != tgtField.Name || srcField.Type != tgtField.Type {
				return false
			}
		}
		return true
	}

	// 对于其他基础类型（int, string 等），直接检查 ConvertibleTo
	return srcType.ConvertibleTo(tgtType)
}
