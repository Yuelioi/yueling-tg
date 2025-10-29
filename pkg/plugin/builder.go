package plugin

import (
	"reflect"
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
func (p *pluginBuilder) Go(parents ...any) Plugin {
	plg := NewBase(p.info)

	// 注册匹配器
	for _, m := range p.matchers {
		plg.AddMatcher(m)
	}

	// 遍历传入的父级
	for _, parent := range parents {
		if parent == nil {
			continue
		}

		// 判断是否实现了 Plugin 接口
		pluginParent, ok := parent.(Plugin)
		if !ok {
			continue
		}

		// 反射拿到结构体指针
		v := reflect.ValueOf(parent)
		if v.Kind() != reflect.Ptr || v.IsNil() {
			continue
		}

		elem := v.Elem()

		// 查找 Base 字段
		field := elem.FieldByName("Base")
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		// 确认类型可以赋值
		if field.Type().AssignableTo(reflect.TypeOf(plg)) {
			field.Set(reflect.ValueOf(plg))
			return pluginParent // ✅ 返回 parent，而不是 plg
		}
	}

	// 如果没传父级或不满足条件，则直接返回 plg
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
