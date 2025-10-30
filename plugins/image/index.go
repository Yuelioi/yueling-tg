package image

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/common"

	"golang.org/x/sync/errgroup"
)

// -------------------- 索引管理 --------------------
func (rg *RandomGenerator) scanAllCategories() (updated bool, err error) {
	var updatedMu sync.Mutex
	updated = false

	g := new(errgroup.Group)

	for _, cat := range rg.config.Categories {
		category := strings.ToLower(cat.Folder) // 避免闭包问题
		g.Go(func() error {
			folder := filepath.Join(rg.config.ImagesFolder, category)
			if _, err := os.Stat(folder); os.IsNotExist(err) {
				return nil
			}
			ok, err := rg.scanFolder(category, folder)
			if err != nil {
				rg.Log.Error().Err(err).Str("category", category).Msg("扫描文件夹失败")
				return err
			}
			if ok {
				updatedMu.Lock()
				updated = true
				updatedMu.Unlock()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return updated, err
	}
	return updated, nil
}

func (rg *RandomGenerator) handleRebuildIndex(ctx *context.Context) {
	rg.Log.Info().Msg("开始重建图片索引，无条件扫描所有图片...")

	// 清空当前索引
	rg.indexDB.mu.Lock()
	rg.indexDB.Images = make(map[string]*ImageIndex)
	rg.indexDB.mu.Unlock()

	// 扫描所有分类
	updated, err := rg.scanAllCategories()
	if err != nil {
		rg.Log.Error().Err(err).Msg("扫描所有图片失败")
		ctx.Replyf("扫描所有图片失败: %v", err)
		return
	}

	if updated {
		rg.Log.Info().Msg("索引重建完成，正在保存索引...")
		if err := rg.saveIndex(); err != nil {
			rg.Log.Error().Err(err).Msg("保存索引失败")
			ctx.Replyf("保存索引失败: %v", err)
			return
		}
		ctx.Reply("索引重建完成 ✅")
		return
	}

	rg.Log.Info().Msg("未发现图片，索引已清空")
	ctx.Reply("未发现图片，索引已清空")
}

// 加载或创建图片索引
func (rg *RandomGenerator) loadOrCreateIndex() error {
	if err := rg.loadIndex(); err != nil {
		rg.Log.Warn().Err(err).Msg("加载索引失败，将扫描所有图片")
	}

	updated, err := rg.scanAllCategories()
	if err != nil {
		return err
	}

	if updated {
		rg.Log.Info().Msg("发现新图片，正在保存索引...")
		return rg.saveIndex()
	}

	rg.Log.Info().Msg("图片索引已是最新 ✅")
	return nil
}

// loadIndex 从文件加载索引
func (rg *RandomGenerator) loadIndex() error {
	data, err := os.ReadFile(rg.config.DBPath)
	if err != nil {
		return err
	}

	rg.indexDB.mu.Lock()
	defer rg.indexDB.mu.Unlock()

	return json.Unmarshal(data, rg.indexDB)
}

// saveIndex 保存索引到文件
func (rg *RandomGenerator) saveIndex() error {
	rg.indexDB.mu.RLock()
	defer rg.indexDB.mu.RUnlock()

	data, err := json.MarshalIndent(rg.indexDB, "", "  ")
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(rg.config.DBPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(rg.config.DBPath, data, 0644)
}

// scanFolder 扫描文件夹，更新索引
func (rg *RandomGenerator) scanFolder(category, folder string) (bool, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return false, err
	}

	g := new(errgroup.Group)
	sem := make(chan struct{}, 8)
	localIndex := make(map[string]*ImageIndex)
	localMu := sync.Mutex{}

	existing := make(map[string]bool)
	rg.indexDB.mu.RLock()
	for _, img := range rg.indexDB.Images {
		existing[img.Path] = true
	}
	rg.indexDB.mu.RUnlock()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" && ext != ".gif" {
			continue
		}

		fullPath := filepath.Join(folder, filename)
		if existing[fullPath] {
			continue
		}

		sem <- struct{}{}
		g.Go(func() error {
			defer func() { <-sem }()
			data, err := os.ReadFile(fullPath)
			if err != nil {
				rg.Log.Error().Err(err).Str("path", fullPath).Msg("读取文件失败")
				return nil
			}

			hash := common.Sha1Hash(data)
			localMu.Lock()
			localIndex[hash] = &ImageIndex{
				Hash:     hash,
				Path:     fullPath,
				Category: category,
				Filename: filename,
			}
			localMu.Unlock()
			rg.Log.Debug().
				Str("hash", hash).
				Str("path", fullPath).
				Msg("添加图片到索引")
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return false, err
	}

	if len(localIndex) == 0 {
		return false, nil
	}

	// 合并
	rg.indexDB.mu.Lock()
	for k, v := range localIndex {
		rg.indexDB.Images[k] = v
	}
	rg.indexDB.mu.Unlock()

	return true, nil
}

// addToIndex 添加单个图片到索引
func (rg *RandomGenerator) addToIndex(path, category, filename string) {
	rg.indexDB.mu.Lock()
	defer rg.indexDB.mu.Unlock()

	rg.indexDB.Images[path] = &ImageIndex{
		Path:     path,
		Category: category,
		Filename: filename,
	}
}

// removeFromIndex 从索引中移除图片
func (rg *RandomGenerator) removeFromIndex(path string) {
	rg.indexDB.mu.Lock()
	defer rg.indexDB.mu.Unlock()

	delete(rg.indexDB.Images, path)
}

// findByHash 根据哈希查找图片
func (rg *RandomGenerator) findHistoryByPath(path string) (*ImageIndex, bool) {
	val, ok := rg.msgHistory.Load(path)
	if !ok {
		return nil, false
	}
	imgIndex, ok := val.(*ImageIndex)
	if !ok {
		return nil, false
	}
	if imgIndex == nil {
		return nil, false
	}

	return imgIndex, ok
}
