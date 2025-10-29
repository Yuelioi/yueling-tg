package core

import (
	"context"
	"sort"

	contextx "yueling_tg/internal/core/context"
	"yueling_tg/internal/middleware"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/handler"
	"yueling_tg/pkg/plugin/provider"

	"github.com/mymmrac/telego"
	"github.com/rs/zerolog"
)

type Runtime struct {
	Api            *telego.Bot
	Logger         zerolog.Logger
	PluginRegistry *plugin.PluginRegistry
	Middlewares    []middleware.Middleware
}

func NewRuntime(api *telego.Bot, logger zerolog.Logger) *Runtime {
	return &Runtime{
		Api:            api,
		Logger:         logger,
		PluginRegistry: plugin.NewPluginRegistry(),
		Middlewares:    []middleware.Middleware{},
	}
}

// Run 启动事件循环
// Run 启动事件循环
func (r *Runtime) Run() {

	gc := handler.InitGlobalContainer()

	// 注册插件列表
	gc.RegisterStatic(
		provider.StaticProvider(func() any {
			return r.PluginRegistry.Plugins()
		}),
	)

	r.Logger.Info().Msg("正在清理历史消息...")

	// 清理历史消息
	r.clearPendingUpdates()

	r.Logger.Info().Msg("Bot 运行中...")

	updates, _ := r.Api.UpdatesViaLongPolling(context.Background(), nil)

	for update := range updates {
		ctx := contextx.NewContext(context.Background(), r.Api, update)

		gc.RegisterDynamic(provider.DynamicProvider(func(ctx *contextx.Context) any {
			return ctx
		}))

		if ctx.GetMessage() != nil {
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

// clearPendingUpdates 清理所有待处理的历史消息
func (r *Runtime) clearPendingUpdates() {
	params := &telego.GetUpdatesParams{
		Offset:  -1,
		Limit:   1,
		Timeout: 0,
	}
	ctx := context.Background()
	updates, err := r.Api.GetUpdates(ctx, params)
	if err != nil {
		r.Logger.Warn().Err(err).Msg("获取历史消息失败")
		return
	}

	if len(updates) > 0 {
		latestID := updates[0].UpdateID
		// 确认所有历史消息
		confirmParams := &telego.GetUpdatesParams{
			Offset:  latestID + 1,
			Limit:   1,
			Timeout: 0,
		}
		r.Api.GetUpdates(ctx, confirmParams)
		r.Logger.Info().Msgf("已跳过 %d 及之前的所有历史消息", latestID)
	} else {
		r.Logger.Info().Msg("没有待处理的历史消息")
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
