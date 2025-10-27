package plugin

import (
	"sort"
	"sync"
	"yueling_tg/core/context"
	"yueling_tg/core/log"

	"github.com/rs/zerolog"
)

type MatcherRegistrar interface {
	RegisterMatchers(ms ...*Matcher)
	UnregisterMatchers(ms *Matcher)
}

var _ MatcherRegistrar = (*MatcherRegistry)(nil)

var (
	once            sync.Once
	matcherRegistry *MatcherRegistry
)

// 匹配器注册中心
type MatcherRegistry struct {
	matchers []*Matcher
	logger   zerolog.Logger
	mu       sync.RWMutex
}

func GetMatcherRegistry() *MatcherRegistry {
	once.Do(func() {
		matcherRegistry = newMatcherRegistry()
	})
	return matcherRegistry
}

func newMatcherRegistry() *MatcherRegistry {
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

// 匹配事件并缓存匹配到的matcher
func (mr *MatcherRegistry) MatchedMatchers(ctx *context.Context) []*Matcher {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	matched := make([]*Matcher, 0, len(mr.matchers))

	for _, m := range mr.matchers {
		if m.Rule.Match(ctx) {
			matched = append(matched, m)
		}
		if m.Block {
			break
		}
	}

	return matched
}
