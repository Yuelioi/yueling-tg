package chat

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// -------------------- 数据持久化 --------------------

func (cp *ChatPlugin) loadPrefs() error {
	data, err := os.ReadFile(cp.config.PrefsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	cp.userPrefs.mu.Lock()
	defer cp.userPrefs.mu.Unlock()

	return json.Unmarshal(data, cp.userPrefs)
}

func (cp *ChatPlugin) savePrefs() error {
	dir := filepath.Dir(cp.config.PrefsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	cp.userPrefs.mu.RLock()
	data, err := json.MarshalIndent(cp.userPrefs, "", "  ")
	cp.userPrefs.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(cp.config.PrefsPath, data, 0644)
}
