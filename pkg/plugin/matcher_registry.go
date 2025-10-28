package plugin

import (
	"sort"
	"sync"
	"yueling_tg/internal/core/log"

	"github.com/rs/zerolog"
)

// 匹配器注册中心
type MatcherRegistry struct {
	matchers []*Matcher
	logger   zerolog.Logger
	mu       sync.RWMutex
}

func NewMatcherRegistry() *MatcherRegistry {
	return &MatcherRegistry{
		matchers: make([]*Matcher, 0),
		logger:   log.NewHandler("MatcherRegistry"),
	}
}

func (mr *MatcherRegistry) RegisterMatchers(ms ...*Matcher) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	mr.matchers = append(mr.matchers, ms...)

	sort.Slice(mr.matchers, func(i, j int) bool {
		return mr.matchers[i].Priority < mr.matchers[j].Priority
	})
}
func (mr *MatcherRegistry) UnregisterMatchers(ms *Matcher) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	for i, m := range mr.matchers {
		if m == ms {
			mr.matchers = append(mr.matchers[:i], mr.matchers[i+1:]...)
			break
		}
	}
}
