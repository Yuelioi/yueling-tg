package plugin

import (
	"yueling_tg/core/log"

	"github.com/rs/zerolog"
)

// plugin/base.go
type Base struct {
	Info     *PluginInfo
	Log      zerolog.Logger
	matchers []*Matcher
}

func NewBase(info *PluginInfo) *Base {
	if info.Extra == nil {
		info.Extra = make(map[string]any)
	}

	return &Base{
		Info:     info,
		Log:      log.NewPlugin(info.ID),
		matchers: make([]*Matcher, 0),
	}
}

func (b *Base) PluginInfo() *PluginInfo {
	return b.Info
}

func (b *Base) Matchers() []*Matcher {
	return b.matchers
}

func (b *Base) SetMatchers(matchers []*Matcher) {
	b.matchers = matchers
}

func (b *Base) AddMatcher(m *Matcher) {
	b.matchers = append(b.matchers, m)
}
