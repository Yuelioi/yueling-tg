package middleware

import (
	"fmt"
	"time"
	"yueling_tg/internal/core/context"
	"yueling_tg/internal/middleware"
)

func TimeoutMiddleware(timeout time.Duration) middleware.Middleware {
	return middleware.MiddlewareFunc("超时中间件", func(ctx *context.Context, next middleware.HandlerFunc) error {

		ec, cancel := ctx.WithTimeout(timeout)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- next(ctx)
		}()

		select {
		case err := <-done:
			return err
		case <-ec.Done():
			return fmt.Errorf("处理超时: %v", timeout)
		}
	})
}
