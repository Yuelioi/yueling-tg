package examples

import "yueling_tg/pkg/plugin"

// -------------------- 配置结构体 --------------------
type PluginConfig2 struct {
	Message1 string `mapstructure:"message_1"` // !配置需要使用mapstructure 因为用的viper 需要通用标签
}

type ExamplePlugin2 struct {
	*plugin.Base
}

func NewExamplePlugin2() plugin.Plugin {
	p := &ExamplePlugin2{}

	// 请勿直接使用插件的日志
	// !⚠️ 会报错!, 因为调用Go(p)后 才会注入BasePlugin, 如果需要初始化, 请在Init中初始化
	p.Log.Debug()

	builder := plugin.New().Go(p)

	return builder
}
