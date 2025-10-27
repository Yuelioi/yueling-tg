package context

import (
	ctx "context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Context 封装了处理 Telegram 消息所需的上下文信息
type Context struct {
	Ctx     ctx.Context
	Api     *tgbotapi.BotAPI
	Update  tgbotapi.Update
	Storage *Storage

	// 元数据
	StartTime   time.Time
	HandlerName string
}

// NewContext 创建新的上下文实例
func NewContext(c ctx.Context, api *tgbotapi.BotAPI, update tgbotapi.Update) *Context {
	return &Context{
		Ctx:       c,
		Api:       api,
		Update:    update,
		Storage:   NewStorage(),
		StartTime: time.Now(),
	}
}

// ============ 链式调用支持 ============

// WithValue 在上下文中添加值（链式）
func (c *Context) WithValue(key string, value any) *Context {
	c.Set(key, value)
	return c
}

// WithTimeout 设置超时时间
func (c *Context) WithTimeout(timeout time.Duration) (*Context, ctx.CancelFunc) {
	newCtx, cancel := ctx.WithTimeout(c.Ctx, timeout)
	c.Ctx = newCtx
	return c, cancel
}

// Deadline 返回上下文的截止时间
func (c *Context) Deadline() (time.Time, bool) {
	return c.Ctx.Deadline()
}

// Done 返回上下文的 done channel
func (c *Context) Done() <-chan struct{} {
	return c.Ctx.Done()
}

// Err 返回上下文的错误
func (c *Context) Err() error {
	return c.Ctx.Err()
}

// Value 从上下文中获取值
func (c *Context) Value(key any) any {
	return c.Ctx.Value(key)
}
