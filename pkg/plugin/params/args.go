package params

import (
	"strings"
)

// CommandArgs 表示命令参数列表
type CommandArgs []string

// Get 获取指定索引的参数，如果索引越界返回空字符串
func (ca CommandArgs) Get(index int) string {
	if index < 0 || index >= len(ca) {
		return ""
	}
	return ca[index]
}

// Len 返回参数数量
func (ca CommandArgs) Len() int {
	return len(ca)
}

// Join 将所有参数用指定分隔符连接
func (ca CommandArgs) Join(sep string) string {
	return strings.Join(ca, sep)
}

// CommandContext 统一的命令上下文结构体
type CommandContext struct {
	// Command 匹配到的命令（不包含 / 前缀）
	Command string

	// Args 匹配到的参数列表
	Args CommandArgs

	// RawText 原始消息文本
	RawText string

	// RawCommand 原始命令字符串（包含 / 前缀）
	RawCommand string
}

// NewCommandContext 从消息文本创建命令上下文
func NewCommandContext(text string) *CommandContext {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return &CommandContext{
			RawText: text,
		}
	}

	rawCmd := parts[0]
	cmd := strings.TrimPrefix(rawCmd, "/")

	// 处理 command@botname 格式
	if idx := strings.Index(cmd, "@"); idx != -1 {
		cmd = cmd[:idx]
	}

	args := make(CommandArgs, 0, len(parts)-1)
	for i := 1; i < len(parts); i++ {
		args = append(args, parts[i])
	}

	return &CommandContext{
		Command:    cmd,
		Args:       args,
		RawText:    text,
		RawCommand: rawCmd,
	}
}

// HasArgs 检查是否有参数
func (ctx *CommandContext) HasArgs() bool {
	return len(ctx.Args) > 0
}

// ArgCount 返回参数数量
func (ctx *CommandContext) ArgCount() int {
	return len(ctx.Args)
}
