package plugin

import (
	"yueling_tg/core/condition"
	"yueling_tg/core/context"
	"yueling_tg/core/handler"
	"yueling_tg/core/permission"
	"yueling_tg/core/provider"
	"yueling_tg/core/rule"
)

type Matcher struct {
	plugin     Plugin                // 插件
	Rule       rule.Rule             // 规则(必须全部满足)
	Permission permission.Permission // 权限(任意满足即可)
	Priority   int                   // 优先级(越大越优先)
	Block      bool                  // 是否阻止事件传播
	Handlers   []*handler.Handler    // 处理器
}

func NewMatcher(rule rule.Rule, handlers ...*handler.Handler) *Matcher {
	return &Matcher{
		Rule:       rule,
		Permission: permission.Everyone(),
		Priority:   10,
		Block:      false,
		Handlers:   handlers,
	}
}

func (m *Matcher) SetPlugin(p Plugin) *Matcher {
	m.plugin = p
	return m
}

func (m *Matcher) Plugin() Plugin {
	return m.plugin
}

func (m *Matcher) SetPriority(priority int) *Matcher {
	m.Priority = priority
	return m
}

func (m *Matcher) SetBlock(block bool) *Matcher {
	m.Block = block
	return m
}

func (m *Matcher) AppendRule(rule rule.Rule) *Matcher {
	m.Rule = condition.All(m.Rule, rule)
	return m
}

func (m *Matcher) AppendPermission(permission permission.Permission) *Matcher {
	m.Permission = condition.Any(m.Permission, permission)
	return m
}

func (m *Matcher) AppendHandler(handler *handler.Handler) *Matcher {
	m.Handlers = append(m.Handlers, handler)
	return m
}

func (m *Matcher) Match(ctx *context.Context) bool {
	if m.Rule != nil && !m.Rule.Match(ctx) {
		return false
	}
	if m.Permission != nil && !m.Permission.Match(ctx) {
		return false
	}
	return true
}

func (m *Matcher) Call(ctx *context.Context, provs ...provider.Provider) error {
	for _, h := range m.Handlers {
		if err := h.Call(ctx, provs...); err != nil {
			return err
		}
	}
	return nil
}
