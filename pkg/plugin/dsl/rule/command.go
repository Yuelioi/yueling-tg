package rule

import (
	"yueling_tg/internal/core/context"

	"regexp"
	"strings"
)

// StartsWith 前缀规则
func StartsWith(prefixes ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		text := ctx.GetMessageText()
		caption := ctx.GetCaption()
		return startsWith(text, prefixes...) || startsWith(caption, prefixes...)
	})
}

// EndsWith 后缀规则
func EndsWith(suffixes ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		text := ctx.GetMessageText()
		caption := ctx.GetCaption()
		return endsWith(text, suffixes...) || endsWith(caption, suffixes...)
	})
}

// FullMatch 完全匹配规则
func FullMatch(patterns ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		text := ctx.GetMessageText()
		caption := ctx.GetCaption()
		return fullMatch(text, patterns...) || fullMatch(caption, patterns...)
	})
}

// Keyword 关键词规则
func Keyword(kwds ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		text := ctx.GetMessageText()
		caption := ctx.GetCaption()
		return keywords(text, kwds...) || keywords(caption, kwds...)
	})
}

// Command 命令规则
func Command(caseSensitive bool, cmds ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		message := strings.TrimSpace(ctx.GetMessageText())
		caption := strings.TrimSpace(ctx.GetCaption())

		if !caseSensitive {
			message = strings.ToLower(message)
			caption = strings.ToLower(caption)
		}

		for _, cmd := range cmds {
			match := cmd
			if !caseSensitive {
				match = strings.ToLower(cmd)
			}
			if strings.HasPrefix(message, match) || strings.HasPrefix(caption, match) {
				return true
			}
		}
		return false
	})
}

// Regex 正则表达式规则
func Regex(patterns ...string) Rule {
	return RuleFunc(func(ctx *context.Context) bool {
		text := ctx.GetMessageText()
		caption := ctx.GetCaption()
		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			if re.MatchString(text) || re.MatchString(caption) {
				return true
			}
		}
		return false
	})
}
