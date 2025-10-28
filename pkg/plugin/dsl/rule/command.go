package rule

import (
	"yueling_tg/internal/core/context"

	"regexp"
	"strings"
)

// StartsWith 前缀规则
func StartsWith(prefix ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		return startsWith(ctx.GetMessageText(), prefix...)
	})
}

// EndsWith 后缀规则
func EndsWith(suffixes ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		return endsWith(ctx.GetMessageText(), suffixes...)
	})
}

func FullMatch(patterns ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		return fullMatch(ctx.GetMessageText(), patterns...)
	})
}

// Keyword 关键词规则
func Keyword(kwds ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		return keywords(ctx.GetMessageText(), kwds...)
	})
}

// Command 命令规则
func Command(caseSensitive bool, cmds ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {

		message := strings.TrimSpace(ctx.GetMessageText())
		if !caseSensitive {
			message = strings.ToLower(message)
		}
		for _, cmd := range cmds {
			match := cmd
			if !caseSensitive {
			}
			if strings.HasPrefix(message, match) {
				return true
			}
		}

		return false
	})
}

// Regex 正则表达式规则
func Regex(patterns ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			if re.MatchString(ctx.GetMessageText()) {
				return true
			}
		}

		return false
	})
}
