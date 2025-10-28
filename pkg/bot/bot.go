package bot

import (
	"yueling_tg/internal/core"
	"yueling_tg/internal/middleware"
	"yueling_tg/pkg/plugin"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

type Bot struct {
	runtime *core.Runtime
}

// NewBot 创建一个新的 Bot 实例
func NewBot(botToken string, logger zerolog.Logger) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	logger.Info().Msgf("授权账户: @%s", api.Self.UserName)
	runtime := core.NewRuntime(api, logger)

	return &Bot{runtime: runtime}, nil
}

// RegisterMiddlewares 注册中间件
func (b *Bot) RegisterMiddlewares(m ...middleware.Middleware) {
	b.runtime.Middlewares = append(b.runtime.Middlewares, m...)
}

// RegisterPlugins 注册插件
func (b *Bot) RegisterPlugins(plugins ...plugin.Plugin) {
	b.runtime.PluginRegistry.RegisterPlugins(plugins...)
}

// Plugins 获取已注册插件
func (b *Bot) Plugins() []plugin.Plugin {
	return b.runtime.PluginRegistry.Plugins()
}

// Run 启动 Bot
func (b *Bot) Run() {
	b.runtime.Run()
}
