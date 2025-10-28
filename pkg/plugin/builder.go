package plugin

import (
	"yueling_tg/pkg/plugin/dsl/condition"
	"yueling_tg/pkg/plugin/dsl/permission"
	"yueling_tg/pkg/plugin/handler"
	"yueling_tg/pkg/plugin/provider"
)

// -----------------------------------------------------------------------------
// Plugin Builder
// -----------------------------------------------------------------------------

type pluginBuilder struct {
	info      *PluginInfo
	matchers  []*Matcher
	providers []provider.Provider
}

// New returns a new plugin builder.
func New() *pluginBuilder {
	return &pluginBuilder{}
}

func (p *pluginBuilder) Info(info *PluginInfo) *pluginBuilder {
	p.info = info
	return p
}

func (p *pluginBuilder) Provide(provider provider.Provider) *pluginBuilder {
	p.providers = append(p.providers, provider)
	return p
}

func (p *pluginBuilder) addMatcher(m *Matcher) *pluginBuilder {
	p.matchers = append(p.matchers, m)
	return p
}

// Build finalizes and returns the plugin instance.
func (p *pluginBuilder) Go() *Base {
	plg := NewBase(p.info)

	for _, m := range p.matchers {
		plg.AddMatcher(m)
	}
	return plg
}

// -----------------------------------------------------------------------------
// Matcher Builder
// -----------------------------------------------------------------------------

type matcherBuilder struct {
	parent      *pluginBuilder
	makeMatcher func(fn any) *Matcher
	perms       []permission.Permission
	priority    int
	block       bool
}

// 创建新的 matcher builder
func newMatcherBuilder(p *pluginBuilder, make func(fn any) *Matcher) *matcherBuilder {
	return &matcherBuilder{parent: p, makeMatcher: make}
}

// 添加权限条件
func (m *matcherBuilder) When(perms ...permission.Permission) *matcherBuilder {
	m.perms = append(m.perms, perms...)
	return m
}

// 设置优先级
func (m *matcherBuilder) Priority(n int) *matcherBuilder {
	m.priority = n
	return m
}

// 是否阻止事件传播
func (m *matcherBuilder) Block(b bool) *matcherBuilder {
	m.block = b
	return m
}

// 设置处理函数并生成 matcher
func (m *matcherBuilder) Do(fn any) *pluginBuilder {
	matcher := m.makeMatcher(fn)
	if len(m.perms) > 0 {
		perm := condition.All(m.perms...)
		matcher = matcher.AppendPermission(perm)
	}

	if m.priority != 0 {
		matcher.Priority = m.priority
	}
	matcher.Block = m.block

	return m.parent.addMatcher(matcher)
}

// -----------------------------------------------------------------------------
// OnXxx DSL functions
// -----------------------------------------------------------------------------

func (p *pluginBuilder) OnFullMatch(patterns ...string) *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		return OnFullMatch(patterns, h)
	})
}

func (p *pluginBuilder) OnKeyword(keywords ...string) *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		return OnKeyword(keywords, h)
	})
}

func (p *pluginBuilder) OnCommand(cmds ...string) *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		h.RegisterDynamicProviders(provider.CommandArgsProvider(cmds), provider.CommandContextProvider(cmds))
		return OnCommand(cmds, true, h)
	})
}

func (p *pluginBuilder) OnCallbackStartsWith(prefixes ...string) *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		h.RegisterDynamicProvider(provider.CallbackDataProvider())
		return OnCallbackStartsWith(prefixes, h)
	})
}

func (p *pluginBuilder) OnCallbackFullMatch(patterns ...string) *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		h.RegisterDynamicProvider(provider.CallbackDataProvider())
		return OnCallbackFullMatch(patterns, h)
	})
}

func (p *pluginBuilder) OnStartsWith(prefixes ...string) *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		return OnStartsWith(prefixes, h)
	})
}

func (p *pluginBuilder) OnEndsWith(suffixes ...string) *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		return OnEndsWith(suffixes, h)
	})
}

func (p *pluginBuilder) OnRegex(patterns ...string) *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		return OnRegex(patterns, h)
	})
}

func (p *pluginBuilder) OnMessage() *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		h.RegisterDynamicProvider(provider.MessageProvider())
		return OnMessage(h)
	})
}

func (p *pluginBuilder) OnNotice() *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		return OnNotice(h)
	})
}

func (p *pluginBuilder) OnCallback() *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		h.RegisterDynamicProvider(provider.CallbackDataProvider())
		return OnCallback(h)
	})
}

func (p *pluginBuilder) OnInlineQuery() *matcherBuilder {
	return newMatcherBuilder(p, func(fn any) *Matcher {
		h := handler.NewHandler(fn)
		h.RegisterDynamicProvider(provider.InlineQueryProvider())
		return OnInlineQuery(h)
	})
}
