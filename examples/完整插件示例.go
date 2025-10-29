package examples

import (
	"context"
	"fmt"

	"yueling_tg/pkg/config"
	"yueling_tg/pkg/plugin"
)

// -------------------- 配置结构体 --------------------
type PluginConfig struct {
	Message1 string `mapstructure:"message_1"` // 打印消息1
	Message2 string `mapstructure:"message_2"` // 打印消息2
}

// -------------------- 插件结构体 --------------------
type ExamplePlugin struct {
	*plugin.Base
	config PluginConfig
}

// -------------------- 初始化插件 --------------------
func (p *ExamplePlugin) Init() error {
	fmt.Println("[ExamplePlugin] 初始化插件")
	return nil
}

// -------------------- 加载插件 --------------------
func (p *ExamplePlugin) Load() error {
	fmt.Println("[ExamplePlugin] 加载插件")
	return nil
}

// -------------------- 验证插件 --------------------
func (p *ExamplePlugin) Validate() error {
	if p.config.Message1 == "" || p.config.Message2 == "" {
		return fmt.Errorf("配置不完整: message_1 或 message_2 为空")
	}
	fmt.Println("[ExamplePlugin] 配置验证通过")
	return nil
}

// -------------------- PluginInfo --------------------
func (p *ExamplePlugin) PluginInfo() *plugin.PluginInfo {
	return p.Base.PluginInfo()
}

// -------------------- Matchers --------------------
func (p *ExamplePlugin) Matchers() []*plugin.Matcher {
	return p.Base.Matchers()
}

// -------------------- 命令处理函数 --------------------
func (p *ExamplePlugin) handlePrint1() {
	fmt.Println("[ExamplePlugin] 打印消息1:", p.config.Message1)
}

func (p *ExamplePlugin) handlePrint2() {
	fmt.Println("[ExamplePlugin] 打印消息2:", p.config.Message2)
}

// -------------------- 构建插件 --------------------
func New() plugin.Plugin {
	// 默认配置
	defaultCfg := PluginConfig{
		Message1: "Hello from Message1",
		Message2: "Hello from Message2",
	}

	ex := &ExamplePlugin{
		config: defaultCfg,
	}

	// 插件信息
	info := &plugin.PluginInfo{
		ID:          "example",
		Name:        "示例插件",
		Description: "演示 Builder + 配置 + 插件接口",
		Version:     "1.0.0",
		Author:      "月离",
		Group:       "示例",
		Extra:       make(map[string]any),
	}

	// 获取配置
	config.GetPluginConfigOrDefault(info.ID, &ex.config, defaultCfg)

	// Builder 模式注册命令
	builder := plugin.New().Info(info)

	// 打印命令1 自动依赖注入
	builder.OnCommand("print1").Do(ex.handlePrint1)

	// 打印命令2 手动调用
	builder.OnCommand("print2").Do(func(c *context.Context) {
		ex.handlePrint2()
	})

	// 返回插件，并注入 Base
	return builder.Go(ex)
}
