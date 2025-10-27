package rule

import (
	"regexp"
	"strings"
)

// StartsWith 判断 msg 是否以任意一个前缀开头（忽略大小写）
func startsWith(msg string, prefixes ...string) bool {
	msg = strings.ToLower(msg)
	for _, p := range prefixes {
		if strings.HasPrefix(msg, strings.ToLower(p)) {
			return true
		}
	}
	return false
}

// EndsWith 判断 msg 是否以任意一个后缀结尾（忽略大小写）
func endsWith(msg string, suffixes ...string) bool {
	msg = strings.ToLower(msg)
	for _, s := range suffixes {
		if strings.HasSuffix(msg, strings.ToLower(s)) {
			return true
		}
	}
	return false
}

// FullMatch 判断 msg 是否完全匹配任意一个字符串（忽略大小写）
func fullMatch(msg string, patterns ...string) bool {
	msg = strings.ToLower(msg)
	for _, p := range patterns {
		if msg == strings.ToLower(p) {
			return true
		}
	}
	return false
}

// Keyword 判断 msg 是否包含任意一个关键字（忽略大小写）
func keywords(msg string, keywords ...string) bool {
	msg = strings.ToLower(msg)
	for _, kw := range keywords {
		if strings.Contains(msg, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// Regex 判断 msg 是否匹配任意一个正则表达式（忽略大小写）
func regex(msg string, patterns ...string) bool {
	msgLower := strings.ToLower(msg)
	for _, pattern := range patterns {
		// 使用 (?i) 前缀忽略大小写
		re := regexp.MustCompile("(?i)" + pattern)
		if re.MatchString(msgLower) {
			return true
		}
	}
	return false
}

// Command 判断 msg 是否以任意命令开头（忽略大小写）
// cmds 可以是 "/start"、"/help" 等
func command(msg string, cmds ...string) bool {
	msg = strings.ToLower(strings.TrimSpace(msg))
	for _, cmd := range cmds {
		if strings.HasPrefix(msg, strings.ToLower(cmd)) {
			return true
		}
	}
	return false
}
