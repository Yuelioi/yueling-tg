package plugin

import (
	"errors"
	"fmt"
	"sync"
	"yueling_tg/internal/core/log"

	"github.com/rs/zerolog"
)

// 插件管理器
type PluginRegistry struct {
	plugins   map[string]Plugin   // 插件映射表，key为插件ID
	pluginMap map[string]Plugin   // 按名称的映射表
	groups    map[string][]Plugin // 分组映射
	logger    zerolog.Logger      // 日志记录器
	mu        sync.RWMutex        // 读写锁，保护并发访问
	mr        *MatcherRegistry    // 匹配器管理器
}

// 创建新的插件管理器实例
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins:   make(map[string]Plugin),
		pluginMap: make(map[string]Plugin),
		groups:    make(map[string][]Plugin),
		logger:    log.NewPluginRegistry("插件管理器"),
		mr:        NewMatcherRegistry(),
	}
}

func (pr *PluginRegistry) Unload() (err error) {
	for _, p := range pr.plugins {
		err = pr.unregisterPlugin(p)
		if err != nil {
			pr.logger.Error().Err(err).Msg("插件卸载失败")
		} else {
			pr.logger.Info().Msg("插件卸载成功")
		}
	}

	return

}

// 根据ID获取插件
func (pr *PluginRegistry) GetPlugin(id string) (Plugin, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	if p, exists := pr.plugins[id]; exists {
		pr.logger.Debug().Str("插件ID", id).Msg("找到插件")
		return p, nil
	}

	pr.logger.Warn().Str("插件ID", id).Msg("未找到插件")
	return nil, fmt.Errorf("未找到插件: %s", id)
}

// 根据分组获取插件
func (pr *PluginRegistry) GetPluginsByGroup(group string) []Plugin {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	plugins := pr.groups[group]
	pr.logger.Debug().Str("分组", group).Int("数量", len(plugins)).Msg("获取分组插件")
	return plugins
}

// 注册插件到管理器
func (pr *PluginRegistry) RegisterPlugins(ps ...Plugin) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	var allMatchers []*Matcher

	for i, p := range ps {
		if p == nil {
			return errors.New("无法注册空插件")
		}

		// 获取插件信息
		metadata := p.PluginInfo()
		if metadata == nil {
			return fmt.Errorf("插件[%d]元数据不能为空", i)
		}

		if metadata.ID == "" {
			return fmt.Errorf("插件[%d]ID不能为空", i)
		}

		if metadata.Name == "" {
			return fmt.Errorf("插件[%d]名称不能为空", i)
		}

		// 检查ID是否已存在
		if _, exists := pr.plugins[metadata.ID]; exists {
			pr.logger.Warn().Str("插件ID", metadata.ID).Msg("存在同ID插件")
			return fmt.Errorf("存在同ID插件: %s", metadata.ID)
		}

		// 检查名称是否已存在
		if _, exists := pr.pluginMap[metadata.Name]; exists {
			pr.logger.Warn().Str("插件名", metadata.Name).Msg("存在同名插件")
			return fmt.Errorf("存在同名插件: %s", metadata.Name)
		}

		// 初始化插件（如果支持）
		if initializer, ok := p.(PluginInitializer); ok {
			if err := initializer.Init(); err != nil {
				pr.logger.Error().
					Err(err).
					Str("插件ID", metadata.ID).
					Msg("插件初始化失败")
				panic(fmt.Errorf("插件[%s]初始化失败: %w", metadata.ID, err))
			}
		}

		// 加载插件（如果支持）
		if loader, ok := p.(PluginLoader); ok {
			if err := loader.Load(); err != nil {
				pr.logger.Error().
					Err(err).
					Str("插件ID", metadata.ID).
					Msg("插件加载失败")
				panic(fmt.Errorf("插件[%s]加载失败: %w", metadata.ID, err))
			}
		}

		// 验证插件配置（如果支持）
		if validator, ok := p.(PluginValidator); ok {
			if err := validator.Validate(); err != nil {
				pr.logger.Error().
					Err(err).
					Str("插件ID", metadata.ID).
					Msg("插件验证失败")
				panic(fmt.Errorf("插件[%s]验证失败: %w", metadata.ID, err))
			}
		}

		// 注册插件
		pr.plugins[metadata.ID] = p
		pr.pluginMap[metadata.Name] = p

		// 按分组索引
		if metadata.Group != "" {
			pr.groups[metadata.Group] = append(pr.groups[metadata.Group], p)
		}

		pr.logger.Info().
			Str("插件ID", metadata.ID).
			Str("插件名", metadata.Name).
			Str("版本", metadata.Version).
			Str("分组", metadata.Group).
			Msg("插件注册成功")

		// 收集匹配器
		matchers := p.Matchers()
		for _, m := range matchers {
			m.SetPlugin(p) // 设置匹配器所属插件
		}
		allMatchers = append(allMatchers, matchers...)
	}

	// 批量注册匹配器
	if len(allMatchers) > 0 {
		pr.mr.RegisterMatchers(allMatchers...)
		pr.logger.Info().Int("匹配器数量", len(allMatchers)).Msg("批量注册匹配器完成")
	}

	return nil
}

// 返回所有已注册插件的副本
func (pr *PluginRegistry) Plugins() []Plugin {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	plugins := make([]Plugin, 0, len(pr.plugins))
	for _, p := range pr.plugins {
		plugins = append(plugins, p)
	}

	return plugins
}

// 根据ID注销插件
func (pr *PluginRegistry) UnregisterPlugin(id string) error {
	if id == "" {
		return fmt.Errorf("插件ID不能为空")
	}

	pr.mu.Lock()
	defer pr.mu.Unlock()

	plugin, exists := pr.plugins[id]
	if !exists {
		pr.logger.Warn().Str("插件ID", id).Msg("尝试注销不存在的插件")
		return fmt.Errorf("未找到插件: %s", id)
	}

	return pr.unregisterPlugin(plugin)
}

// unregisterPlugin 内部注销插件方法
func (pr *PluginRegistry) unregisterPlugin(plugin Plugin) error {
	metadata := plugin.PluginInfo()

	// 卸载插件（如果支持）
	if unloader, ok := plugin.(PluginUnloader); ok {
		if err := unloader.Unload(); err != nil {
			pr.logger.Error().
				Err(err).
				Str("插件ID", metadata.ID).
				Msg("插件卸载失败")
			// 继续注销，但记录错误
		}
	}

	// 从映射表中删除
	delete(pr.plugins, metadata.ID)
	delete(pr.pluginMap, metadata.Name)

	// 从分组中删除
	if metadata.Group != "" {
		groupPlugins := pr.groups[metadata.Group]
		for i, p := range groupPlugins {
			if p.PluginInfo().ID == metadata.ID {
				pr.groups[metadata.Group] = append(groupPlugins[:i], groupPlugins[i+1:]...)
				break
			}
		}
	}

	// TODO: 从匹配器注册表中删除相关匹配器

	pr.logger.Info().
		Str("插件ID", metadata.ID).
		Str("插件名", metadata.Name).
		Msg("插件注销成功")

	return nil
}

// 配置插件
func (pr *PluginRegistry) ConfigurePlugin(id string, config map[string]any) error {
	plugin, err := pr.GetPlugin(id)
	if err != nil {
		return err
	}

	configurable, ok := plugin.(PluginConfigurable)
	if !ok {
		return fmt.Errorf("插件[%s]不支持配置", id)
	}

	if err := configurable.SetConfig(config); err != nil {
		return fmt.Errorf("配置插件[%s]失败: %w", id, err)
	}

	// 验证配置
	if validator, ok := plugin.(PluginValidator); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("插件[%s]配置验证失败: %w", id, err)
		}
	}

	pr.logger.Info().Str("插件ID", id).Msg("插件配置成功")
	return nil
}

// 对所有插件进行健康检查
func (pr *PluginRegistry) HealthCheck() map[string]error {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	results := make(map[string]error)

	for id, plugin := range pr.plugins {
		if checker, ok := plugin.(PluginHealthChecker); ok {
			results[id] = checker.HealthCheck()
		}
	}

	pr.logger.Debug().Int("检查数量", len(results)).Msg("健康检查完成")
	return results
}
