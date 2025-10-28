package provider

import (
	"strings"
	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin/params"
)

// CommandContextProvider 提供完整的命令上下文
func CommandContextProvider(cmds []string) Provider {
	return DynamicProvider(func(ctx *context.Context) any {
		msg := ctx.GetMessage()
		if msg == nil || msg.Text == "" {
			return params.CommandContext{
				RawText: "",
			}
		}

		text := msg.Text
		matchedCmd := ""

		// 找到匹配的命令
		for _, cmd := range cmds {
			if strings.HasPrefix(text, cmd) {
				matchedCmd = cmd
				break
			}
		}

		// 如果没有匹配到命令，返回空上下文
		if matchedCmd == "" {
			return params.CommandContext{
				RawText: text,
			}
		}

		// 移除命令前缀，获取参数部分
		argsText := strings.TrimPrefix(text, matchedCmd)
		argsText = strings.TrimSpace(argsText)

		// 按空格和换行分割参数
		parts := strings.FieldsFunc(argsText, func(r rune) bool {
			return r == ' ' || r == '\n' || r == '\r' || r == '\t'
		})

		var args params.CommandArgs
		for _, p := range parts {
			if p != "" {
				args = append(args, p)
			}
		}

		// 提取纯命令名（去除 / 和 @botname）
		cmdName := strings.TrimPrefix(matchedCmd, "/")
		if idx := strings.Index(cmdName, "@"); idx != -1 {
			cmdName = cmdName[:idx]
		}

		return params.CommandContext{
			Command:    cmdName,
			Args:       args,
			RawText:    text,
			RawCommand: matchedCmd,
		}
	})
}

// CommandArgsProvider 提供命令参数
func CommandArgsProvider(cmds []string) Provider {
	return DynamicProvider(func(ctx *context.Context) any {
		cmdCtx := CommandContextProvider(cmds).Provide(ctx).(params.CommandContext)
		return cmdCtx.Args
	})
}

func MessageProvider() Provider {
	return DynamicProvider(func(ctx *context.Context) any {
		return ctx.GetMessage()
	})
}

func CallbackDataProvider() Provider {
	return DynamicProvider(func(ctx *context.Context) any {
		return ctx.GetCallbackData()
	})
}

func InlineQueryProvider() Provider {
	return DynamicProvider(func(ctx *context.Context) any {
		return ctx.GetInlineQuery()
	})
}
