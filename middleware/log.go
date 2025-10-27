package middleware

import (
	"time"
	"yueling_tg/core/context"
	"yueling_tg/core/log"
	"yueling_tg/core/middleware"
)

var logger = log.NewMiddleware("事件耗时统计")

// LoggingMiddleware 日志中间件
func LoggingMiddleware() middleware.Middleware {
	return middleware.MiddlewareFunc("日志中间件", func(ctx *context.Context, next middleware.HandlerFunc) error {
		start := time.Now()

		err := next(ctx)

		duration := time.Since(start)
		if err != nil {
			return err

		} else {

			pluginName, ok := ctx.Storage.Get(context.PluginName)
			if ok {
				logger.Info().Msgf("事件处理成功 BOT: %v 耗时: %v", pluginName, duration)
			} else {
				logger.Info().Msgf("事件处理成功 BOT: %v  耗时: %v", "未知插件", duration)
			}
		}

		return err
	})
}
