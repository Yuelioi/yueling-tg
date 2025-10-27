package plugin

// 插件接口
type Plugin interface {
	// 插件信息
	PluginInfo() *PluginInfo

	// 获取匹配器机器人的核心功能
	Matchers() []*Matcher
}

// 需要初始化的插件
type PluginInitializer interface {
	Init() error
}

// 需要加载的插件
type PluginLoader interface {
	Load() error
}

// 需要卸载的插件
type PluginUnloader interface {
	Unload() error
}

// 可配置的插件
type PluginConfigurable interface {
	SetConfig(config map[string]any) error
	GetConfig() map[string]any
}

// 支持验证的插件
type PluginValidator interface {
	Validate() error
}

// 支持健康检查的插件
type PluginHealthChecker interface {
	HealthCheck() error
}
