package bot

import (
	"sort"
	"yueling_tg/core/middleware"
	"yueling_tg/core/plugin"

	"yueling_tg/core/context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"

	ctxx "context"
)

var b *Bot

type Bot struct {
	Api         *tgbotapi.BotAPI
	logger      zerolog.Logger
	plugins     []plugin.Plugin
	middlewares []middleware.Middleware
}

func NewBot(botToken string, logger zerolog.Logger) *Bot {

	botApi, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logger.Panic().Err(err).Msg("创建 Bot 失败")
	}

	logger.Info().Msgf("授权账户: @%s", botApi.Self.UserName)

	b = &Bot{
		Api:    botApi,
		logger: logger,
	}
	return b

}

func Plugins() []plugin.Plugin {
	return b.plugins
}

func (b *Bot) RegisterMiddlewares(middlewares ...middleware.Middleware) {
	b.middlewares = append(b.middlewares, middlewares...)
}

func (b *Bot) RegisterPlugins(plugins ...plugin.Plugin) {
	for _, p := range plugins {
		b.logger.Info().Msgf("注册 %s 插件", p.PluginInfo().Name)

		for _, matcher := range p.Matchers() {
			matcher.SetPlugin(p)
		}
	}
	b.plugins = append(b.plugins, plugins...)
}
func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// ---------------------
	// 丢弃历史消息
	updates, _ := b.Api.GetUpdates(tgbotapi.UpdateConfig{
		Offset: 0,
		Limit:  1,
	})
	if len(updates) > 0 {
		lastUpdateID := updates[len(updates)-1].UpdateID
		u.Offset = lastUpdateID + 1
	}

	b.logger.Info().Msg("运行成功")

	updatesChan := b.Api.GetUpdatesChan(u)

	for update := range updatesChan {

		ctx := context.NewContext(ctxx.Background(), b.Api, update)

		if ctx.GetMessage() == nil {
			b.logger.Info().
				Str("user", ctx.GetUsername()).
				Str("text", ctx.GetMessageText()).
				Msg("收到消息")
		}

		handler := middleware.Chain(b.middlewares, func(ctx *context.Context) error {
			return b.processMatchers(ctx)
		})

		if err := handler(ctx); err != nil {
			b.logger.Error().Err(err).Msg("处理消息失败")
		}
	}
}

func (b *Bot) processMatchers(ctx *context.Context) error {
	// 收集所有匹配器
	var allMatchers []*plugin.Matcher
	for _, p := range b.plugins {
		matchers := p.Matchers()
		allMatchers = append(allMatchers, matchers...)
	}

	// 按优先级排序（优先级高的在前）
	sort.Slice(allMatchers, func(i, j int) bool {
		return allMatchers[i].Priority > allMatchers[j].Priority
	})

	// 依次匹配并执行
	for _, matcher := range allMatchers {
		// 检查规则和权限
		if !matcher.Match(ctx) {
			continue
		}

		// 匹配成功，记录日志
		pluginName := "unknown"
		if matcher.Plugin() != nil {
			pluginName = matcher.Plugin().PluginInfo().Name
		}

		b.logger.Debug().
			Str("plugin", pluginName).
			Int("priority", matcher.Priority).
			Msg("匹配成功")

		ctx.Storage.Set(context.PluginName, pluginName)

		// 执行处理器
		if err := matcher.Call(ctx); err != nil {
			b.logger.Error().
				Err(err).
				Str("plugin", pluginName).
				Msg("执行处理器失败")

			// 继续处理下一个匹配器（除非需要中断）
			if !matcher.Block {
				continue
			}
		}

		// 如果设置了阻止事件传播，则停止处理
		if matcher.Block {
			b.logger.Debug().
				Str("plugin", pluginName).
				Msg("事件传播被阻止")
			break
		}
	}

	return nil
}
