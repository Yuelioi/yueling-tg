package context

import (
	"sync"
)

// Storage 线程安全的存储结构
type Storage struct {
	data map[string]any
	mu   sync.RWMutex
}

// NewStorage 创建新的存储实例
func NewStorage() *Storage {
	return &Storage{
		data: make(map[string]any),
	}
}

// Set 设置值
func (s *Storage) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get 获取值
func (s *Storage) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// MustGet 获取值，不存在则 panic
func (s *Storage) MustGet(key string) any {
	if val, ok := s.Get(key); ok {
		return val
	}
	panic("key '" + key + "' does not exist")
}

// GetString 获取字符串类型的值
func (s *Storage) GetString(key string) (string, bool) {
	if val, ok := s.Get(key); ok {
		str, ok := val.(string)
		return str, ok
	}
	return "", false
}

// GetInt 获取整数类型的值
func (s *Storage) GetInt(key string) (int, bool) {
	if val, ok := s.Get(key); ok {
		i, ok := val.(int)
		return i, ok
	}
	return 0, false
}

// GetInt64 获取 int64 类型的值
func (s *Storage) GetInt64(key string) (int64, bool) {
	if val, ok := s.Get(key); ok {
		i, ok := val.(int64)
		return i, ok
	}
	return 0, false
}

// GetBool 获取布尔类型的值
func (s *Storage) GetBool(key string) (bool, bool) {
	if val, ok := s.Get(key); ok {
		b, ok := val.(bool)
		return b, ok
	}
	return false, false
}

// Delete 删除值
func (s *Storage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// Clear 清空所有数据
func (s *Storage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]any)
}

// ============ Storage 便捷方法 ============

// Set 设置存储值
func (c *Context) Set(key string, value any) {
	c.Storage.Set(key, value)
}

// Get 获取存储值
func (c *Context) Get(key string) (any, bool) {
	return c.Storage.Get(key)
}

// MustGet 获取存储值，不存在则 panic
func (c *Context) MustGet(key string) any {
	return c.Storage.MustGet(key)
}

// GetString 获取字符串类型的值
func (c *Context) GetString(key string) (string, bool) {
	return c.Storage.GetString(key)
}

// GetInt 获取整数类型的值
func (c *Context) GetInt(key string) (int, bool) {
	return c.Storage.GetInt(key)
}

// GetInt64 获取 int64 类型的值
func (c *Context) GetInt64(key string) (int64, bool) {
	return c.Storage.GetInt64(key)
}

// GetBool 获取布尔类型的值
func (c *Context) GetBool(key string) (bool, bool) {
	return c.Storage.GetBool(key)
}

// Delete 删除存储值
func (c *Context) Delete(key string) {
	c.Storage.Delete(key)
}
