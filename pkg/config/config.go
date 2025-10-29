package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// -------------------- 配置管理器 --------------------

// ConfigManager 全局配置管理器
type ConfigManager struct {
	viper  *viper.Viper
	mu     sync.RWMutex
	path   string
	format string // json, toml, yaml 等
}

var (
	globalManager *ConfigManager
	once          sync.Once
)

// InitConfigManager 初始化全局配置管理器（在bot启动时调用）
// configPath: 配置文件路径，如 "./config/plugins.json" 或 "./config/plugins.toml"
func InitConfigManager(configPath string) error {
	var err error
	once.Do(func() {
		v := viper.New()

		// 自动检测文件格式
		ext := filepath.Ext(configPath)
		if len(ext) > 0 {
			ext = ext[1:] // 去掉点号
		} else {
			ext = "json" // 默认格式
		}

		v.SetConfigFile(configPath)
		v.SetConfigType(ext)

		globalManager = &ConfigManager{
			viper:  v,
			path:   configPath,
			format: ext,
		}

		// 尝试读取配置文件
		if err = v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// 配置文件不存在，创建默认配置
				if err = globalManager.ensureConfigDir(); err != nil {
					return
				}
				err = globalManager.save()
				return
			}
			err = fmt.Errorf("读取配置文件失败: %w", err)
		}
	})

	return err
}

// GetManager 获取全局配置管理器实例
func GetManager() *ConfigManager {
	if globalManager == nil {
		panic("配置管理器未初始化，请先调用 InitConfigManager")
	}
	return globalManager
}

// ensureConfigDir 确保配置文件目录存在
func (cm *ConfigManager) ensureConfigDir() error {
	dir := filepath.Dir(cm.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	return nil
}

// save 保存配置到文件
func (cm *ConfigManager) save() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if err := cm.ensureConfigDir(); err != nil {
		return err
	}

	if err := cm.viper.WriteConfig(); err != nil {
		// 如果文件不存在，使用 SafeWriteConfig
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := cm.viper.SafeWriteConfig(); err != nil {
				return fmt.Errorf("保存配置文件失败: %w", err)
			}
			return nil
		}
		return fmt.Errorf("保存配置文件失败: %w", err)
	}
	return nil
}

// Get 获取指定插件的配置（返回 map[string]interface{}）
func (cm *ConfigManager) Get(pluginID string) (map[string]interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	key := fmt.Sprintf("plugins.%s", pluginID)
	if !cm.viper.IsSet(key) {
		return nil, false
	}

	cfg := cm.viper.GetStringMap(key)
	return cfg, true
}

// GetRaw 获取指定插件的原始配置（返回任意类型）
func (cm *ConfigManager) GetRaw(pluginID string) (interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	key := fmt.Sprintf("plugins.%s", pluginID)
	if !cm.viper.IsSet(key) {
		return nil, false
	}

	return cm.viper.Get(key), true
}

// Set 设置指定插件的配置
func (cm *ConfigManager) Set(pluginID string, config interface{}) error {
	cm.mu.Lock()
	key := fmt.Sprintf("plugins.%s", pluginID)
	cm.viper.Set(key, config)
	cm.mu.Unlock()

	return cm.save()
}

// Delete 删除指定插件的配置
func (cm *ConfigManager) Delete(pluginID string) error {
	cm.mu.Lock()

	// viper 没有直接的删除方法，需要重建配置
	plugins := cm.viper.GetStringMap("plugins")
	delete(plugins, pluginID)
	cm.viper.Set("plugins", plugins)

	cm.mu.Unlock()

	return cm.save()
}

// Exists 检查插件配置是否存在
func (cm *ConfigManager) Exists(pluginID string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	prefix := fmt.Sprintf("plugins.%s", pluginID)
	for _, key := range cm.viper.AllKeys() {
		if key == prefix || strings.HasPrefix(key, prefix+".") {
			return true
		}
	}
	return false
}

// GetAllPluginIDs 获取所有已配置的插件ID列表
func (cm *ConfigManager) GetAllPluginIDs() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	plugins := cm.viper.GetStringMap("plugins")
	ids := make([]string, 0, len(plugins))
	for id := range plugins {
		ids = append(ids, id)
	}
	return ids
}

// Reload 重新加载配置文件
func (cm *ConfigManager) Reload() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if err := cm.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("重新加载配置文件失败: %w", err)
	}
	return nil
}

// -------------------- 插件配置辅助函数 --------------------

// GetPluginConfig 插件获取并解析自己的配置
func GetPluginConfig(pluginID string, target interface{}) error {
	manager := GetManager()

	raw, exists := manager.GetRaw(pluginID)
	if !exists {
		return fmt.Errorf("插件 %s 的配置不存在", pluginID)
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "mapstructure",
		Result:  target,
	})
	if err != nil {
		return fmt.Errorf("创建解码器失败: %w", err)
	}

	if err := decoder.Decode(raw); err != nil {
		return fmt.Errorf("解析插件 %s 的配置失败: %w", pluginID, err)
	}

	return nil
}

// GetPluginConfigOrDefault 获取插件配置，如果不存在则使用默认值并保存
func GetPluginConfigOrDefault(pluginID string, target interface{}, defaultConfig interface{}) error {
	manager := GetManager()

	if !manager.Exists(pluginID) {
		// 配置不存在，使用默认值并保存
		if err := manager.Set(pluginID, defaultConfig); err != nil {
			return fmt.Errorf("保存插件 %s 的默认配置失败: %w", pluginID, err)
		}

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			TagName: "mapstructure",
			Result:  target,
		})
		if err != nil {
			return fmt.Errorf("创建解码器失败: %w", err)
		}

		if err := decoder.Decode(defaultConfig); err != nil {
			return fmt.Errorf("解析默认配置失败: %w", err)
		}

		return nil
	}

	// 配置存在，直接获取并解析
	return GetPluginConfig(pluginID, target)
}

// SetPluginConfig 插件设置自己的配置
func SetPluginConfig(pluginID string, config interface{}) error {
	manager := GetManager()
	return manager.Set(pluginID, config)
}

// DeletePluginConfig 删除插件配置
func DeletePluginConfig(pluginID string) error {
	manager := GetManager()
	return manager.Delete(pluginID)
}

// PluginConfigExists 检查插件配置是否存在
func PluginConfigExists(pluginID string) bool {
	manager := GetManager()
	return manager.Exists(pluginID)
}

// ReloadConfig 重新加载配置文件
func ReloadConfig() error {
	manager := GetManager()
	return manager.Reload()
}

// GetAllPlugins 获取所有已配置的插件ID
func GetAllPlugins() []string {
	manager := GetManager()
	return manager.GetAllPluginIDs()
}

// GetString 获取插件配置中的字符串值
func GetString(pluginID, key string, defaultValue string) string {
	manager := GetManager()
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	fullKey := fmt.Sprintf("plugins.%s.%s", pluginID, key)
	if !manager.viper.IsSet(fullKey) {
		return defaultValue
	}
	return manager.viper.GetString(fullKey)
}

// GetInt 获取插件配置中的整数值
func GetInt(pluginID, key string, defaultValue int) int {
	manager := GetManager()
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	fullKey := fmt.Sprintf("plugins.%s.%s", pluginID, key)
	if !manager.viper.IsSet(fullKey) {
		return defaultValue
	}
	return manager.viper.GetInt(fullKey)
}

// GetBool 获取插件配置中的布尔值
func GetBool(pluginID, key string, defaultValue bool) bool {
	manager := GetManager()
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	fullKey := fmt.Sprintf("plugins.%s.%s", pluginID, key)
	if !manager.viper.IsSet(fullKey) {
		return defaultValue
	}
	return manager.viper.GetBool(fullKey)
}

// GetStringSlice 获取插件配置中的字符串数组
func GetStringSlice(pluginID, key string, defaultValue []string) []string {
	manager := GetManager()
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	fullKey := fmt.Sprintf("plugins.%s.%s", pluginID, key)
	if !manager.viper.IsSet(fullKey) {
		return defaultValue
	}
	return manager.viper.GetStringSlice(fullKey)
}
