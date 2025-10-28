package permission

import (
	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin/dsl/condition"
)

// 权限
type Permission = condition.Condition

// PermissionFunc 权限函数类型
type PermissionFunc func(ctx *context.Context) bool

func (f PermissionFunc) Match(ctx *context.Context) bool {
	return f(ctx)
}
