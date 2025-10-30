package bot

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"yueling_tg/internal/core"
	logx "yueling_tg/internal/core/log"
	"yueling_tg/internal/middleware"
	"yueling_tg/pkg/config"
	"yueling_tg/pkg/plugin"

	"github.com/mymmrac/telego"
	"github.com/rs/zerolog/log"
)

type Bot struct {
	runtime *core.Runtime
}

type ZerologWrapper struct{}

func (z ZerologWrapper) Debugf(format string, args ...any) {
	if strings.Contains(format, "API call to") ||
		strings.Contains(format, "API response") {
		return
	}

	log.Debug().Msg(format)
	log.Debug().Msgf(format, args...)
}

// Errorf logs error messages
func (z ZerologWrapper) Errorf(format string, args ...any) {

	msg := fmt.Sprintf(format, args...)

	if strings.Contains(format, "Retrying getting") ||
		strings.Contains(format, "Getting updates") ||
		strings.Contains(msg, "context deadline exceeded") {
		return
	}

	log.Error().Msgf(format, args...)
}

// 创建一个新的 Bot 实例
func NewBot(botToken, configPath string, client *http.Client) (*Bot, error) {
	loggerWrapper := ZerologWrapper{}

	bot, err := telego.NewBot(botToken,
		telego.WithDefaultDebugLogger(),
		telego.WithHTTPClient(client),
		telego.WithLogger(loggerWrapper),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	config.InitConfigManager(configPath)

	b, err := bot.GetMe(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fullName := b.FirstName + b.LastName

	botLogger := logx.NewBot(fullName)
	botLogger.Info().Msgf("授权账户: @%s", fullName)

	runtime := core.NewRuntime(bot, botLogger)

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
