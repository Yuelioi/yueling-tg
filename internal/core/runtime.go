package core

import (
	"context"
	"sort"

	contextx "yueling_tg/internal/core/context"
	"yueling_tg/internal/middleware"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/handler"
	"yueling_tg/pkg/plugin/provider"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
)

type Runtime struct {
	Api            *tgbotapi.BotAPI
	Logger         zerolog.Logger
	PluginRegistry *plugin.PluginRegistry
	Middlewares    []middleware.Middleware
}

func NewRuntime(api *tgbotapi.BotAPI, logger zerolog.Logger) *Runtime {
	return &Runtime{
		Api:            api,
		Logger:         logger,
		PluginRegistry: plugin.NewPluginRegistry(),
		Middlewares:    []middleware.Middleware{},
	}
}

// Run 启动事件循环
func (r *Runtime) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := r.Api.GetUpdates(tgbotapi.UpdateConfig{Offset: 0, Limit: 1})
	if len(updates) > 0 {
		u.Offset = updates[len(updates)-1].UpdateID + 1
	}

	r.Logger.Info().Msg("Bot 运行中...")

	gc := handler.InitGlobalContainer()

	// 注册插件列表
	gc.RegisterStatic(
		provider.StaticProvider(func() any {
			return r.PluginRegistry.Plugins()
		}),
	)

	updatesChan := r.Api.GetUpdatesChan(u)

	for update := range updatesChan {
		ctx := contextx.NewContext(context.Background(), r.Api, update)

		gc.RegisterDynamic(provider.DynamicProvider(func(ctx *contextx.Context) any {
			return ctx
		}))

		if ctx.GetMessage() == nil {
			r.Logger.Info().
				Str("user", ctx.GetUsername()).
				Str("text", ctx.GetMessageText()).
				Msg("收到消息")
		}

		handler := middleware.Chain(r.Middlewares, func(ctx *contextx.Context) error {
			return r.processMatchers(ctx)
		})

		if err := handler(ctx); err != nil {
			r.Logger.Error().Err(err).Msg("处理消息失败")
		}
	}
}

func (r *Runtime) processMatchers(ctx *contextx.Context) error {
	var allMatchers []*plugin.Matcher
	for _, p := range r.PluginRegistry.Plugins() {
		allMatchers = append(allMatchers, p.Matchers()...)
	}

	sort.Slice(allMatchers, func(i, j int) bool {
		return allMatchers[i].Priority > allMatchers[j].Priority
	})

	for _, matcher := range allMatchers {
		if !matcher.Match(ctx) {
			continue
		}

		pluginName := "unknown"
		if matcher.Plugin() != nil {
			pluginName = matcher.Plugin().PluginInfo().Name
		}

		r.Logger.Debug().
			Str("plugin", pluginName).
			Int("priority", matcher.Priority).
			Msg("匹配成功")

		ctx.Storage.Set(contextx.PluginName, pluginName)

		if err := matcher.Call(ctx); err != nil {
			r.Logger.Error().Err(err).
				Str("plugin", pluginName).
				Msg("执行处理器失败")

			if !matcher.Block {
				continue
			}
		}

		if matcher.Block {
			r.Logger.Debug().
				Str("plugin", pluginName).
				Msg("事件传播被阻止")
			break
		}
	}
	return nil
}
