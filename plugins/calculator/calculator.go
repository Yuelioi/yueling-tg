package calculator

import (
	"strings"

	"github.com/Knetic/govaluate"

	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/params"
)

var _ plugin.Plugin = (*CalculatorPlugin)(nil)

// 插件结构
type CalculatorPlugin struct {
	*plugin.Base
}

func New() plugin.Plugin {
	info := &plugin.PluginInfo{
		ID:          "calculator",
		Name:        "计算器",
		Description: "支持加减乘除、比较、位运算、安全计算",
		Version:     "1.0.0",
		Author:      "月离",
		Usage:       "计算 <表达式>\n示例：计算 12*21 + 5",
		Extra: map[string]any{
			"group":    "工具",
			"commands": []string{"计算"},
		},
		Group: "工具",
	}
	cp := &CalculatorPlugin{}

	return plugin.New().Info(info).OnCommand("计算").Block(true).Do(cp.calcHandler).Go(cp)
}

// 处理器
func (cp *CalculatorPlugin) calcHandler(ctx *context.Context, commandArgs params.CommandArgs) {
	if len(commandArgs) == 0 {
		ctx.Reply("⚠ 请提供需要计算的表达式，例如：12*21")
		return
	}

	// 拼接表达式
	exp := strings.Join(commandArgs, "")
	exp = strings.ReplaceAll(exp, " ", "") // 去掉空格

	// 安全解析表达式
	expression, err := govaluate.NewEvaluableExpression(exp)
	if err != nil {
		ctx.Reply("❌ 语法错误: " + err.Error())
		return
	}

	// 执行计算
	result, err := expression.Evaluate(nil)
	if err != nil {
		ctx.Reply("❌ 计算错误: " + err.Error())
		return
	}

	// 返回结果
	ctx.Replyf("🧮 计算结果: %v", result)
	cp.Log.Info().Str("expression", exp).Msg("计算完成")
}
