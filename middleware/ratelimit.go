package middleware

import (
	"fmt"
	"sync"
	"time"
	"yueling_tg/core/context"
	"yueling_tg/core/middleware"
)

// RateLimitMiddleware 频率限制中间件
func RateLimitMiddleware(maxRequests int, window time.Duration) middleware.Middleware {
	type userLimit struct {
		requests []time.Time
		mu       sync.Mutex
	}

	users := sync.Map{}

	return middleware.MiddlewareFunc("频率限制中间件", func(ctx *context.Context, next middleware.HandlerFunc) error {
		userID := ctx.GetUserID()

		now := time.Now()

		// 获取或创建用户限制记录
		value, _ := users.LoadOrStore(userID, &userLimit{})
		ul := value.(*userLimit)

		ul.mu.Lock()
		defer ul.mu.Unlock()

		// 清理过期记录
		var validRequests []time.Time
		for _, t := range ul.requests {
			if now.Sub(t) < window {
				validRequests = append(validRequests, t)
			}
		}

		if len(validRequests) >= maxRequests {
			return fmt.Errorf("频率限制: 用户 %s 在 %v 内已发送 %d 条消息", userID, window, maxRequests)
		}

		// 记录本次请求
		validRequests = append(validRequests, now)
		ul.requests = validRequests

		return next(ctx)
	})
}
