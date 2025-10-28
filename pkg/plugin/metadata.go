package plugin

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string         // 插件标识(必填)
	Name        string         // 插件显示名称(必填)
	Description string         // 插件描述
	Version     string         // 插件版本
	Author      string         // 插件作者
	Usage       string         // 插件用途
	Examples    []string       // 插件示例
	Group       string         // 插件分组
	Extra       map[string]any // 额外信息
}
