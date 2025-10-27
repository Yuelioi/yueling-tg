package permission

import (
	"yueling_tg/core/context"

	"yueling_tg/core/condition"
)

// Everyone 所有人权限
func Everyone() Permission {
	return PermissionFunc(func(ctx *context.Context) bool {
		return true
	})
}

// 仅超级用户权限
func SuperUser(superUsers ...int64) Permission {
	userSet := make(map[int64]bool)
	for _, user := range superUsers {
		userSet[user] = true
	}

	return PermissionFunc(func(ctx *context.Context) bool {
		return userSet[ctx.GetUserID()]

	})
}

// GroupOwner 仅群主权限
func GroupOwner() Permission {
	return PermissionFunc(func(ctx *context.Context) bool {
		return getUserRole(ctx) == "owner"
	})
}

// GroupAdmin 仅群管理员权限
func GroupAdmin() Permission {
	return PermissionFunc(func(ctx *context.Context) bool {
		return getUserRole(ctx) == "admin"
	})
}

// GroupMember 仅群成员权限
func GroupMember() Permission {
	return PermissionFunc(func(ctx *context.Context) bool {
		return getUserRole(ctx) == "member"
	})
}

// 超级用户、群主、管理员
func GroupAdminOrOwner() Permission {
	return condition.Any(SuperUser(), GroupOwner(), GroupAdmin())
}
