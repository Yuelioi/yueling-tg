package banword

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/config"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/params"
)

var _ plugin.Plugin = (*BanwordPlugin)(nil)

// -------------------- 数据结构 --------------------

type GroupBanwords struct {
	GroupID  int64    `json:"group_id"`
	Keywords []string `json:"keywords"`
}

type BanwordDB struct {
	Groups map[int64][]string `json:"groups"` // group_id -> keywords
	mu     sync.RWMutex       `json:"-"`
}

// -------------------- 插件结构 --------------------

type PluginConfig struct {
	DBPath string `mapstructure:"db_path"`
}

type BanwordPlugin struct {
	*plugin.Base
	db     *BanwordDB
	config PluginConfig
}

func New() plugin.Plugin {
	bp := &BanwordPlugin{
		db: &BanwordDB{
			Groups: make(map[int64][]string),
		},
	}

	// 插件信息
	info := &plugin.PluginInfo{
		ID:          "banword",
		Name:        "关键词屏蔽",
		Description: "管理群组屏蔽关键词的插件",
		Version:     "1.0.0",
		Author:      "月离",
		Usage:       "添加屏蔽 <关键词1> [关键词2...]\n删除屏蔽 <关键词1> [关键词2...]\n查看屏蔽",
		Group:       "群管",
		Extra:       make(map[string]any),
	}

	// 默认配置
	defaultCfg := PluginConfig{
		DBPath: "./data/banword.json",
	}

	// 加载或创建配置
	if err := config.GetPluginConfigOrDefault(info.ID, &bp.config, defaultCfg); err != nil {
		panic(fmt.Sprintf("加载插件配置失败: %v", err))
	}

	// 初始化 Builder
	builder := plugin.New().Info(info)

	// 消息预处理（最高优先级，用于拦截屏蔽词）
	builder.OnMessage().Priority(100).Do(bp.handleMessageCheck)

	// 管理命令
	builder.OnCommand("添加屏蔽").Priority(10).Do(bp.handleAddBanword)
	builder.OnCommand("删除屏蔽", "取消屏蔽").Priority(10).Do(bp.handleDeleteBanword)
	builder.OnCommand("查看屏蔽").Priority(10).Do(bp.handleListBanword)

	// 返回插件，并注入 Base
	return builder.Go(bp)
}

func (bp *BanwordPlugin) Init() error {
	// 加载数据
	if err := bp.loadData(); err != nil {
		bp.Log.Warn().Msgf("⚠️ 加载屏蔽词数据失败，使用空数据库: %v", err)
	} else {
		totalKeywords := 0
		for _, keywords := range bp.db.Groups {
			totalKeywords += len(keywords)
		}
		bp.Log.Info().Msgf("已加载 %d 个群组的屏蔽词，共 %d 个关键词", len(bp.db.Groups), totalKeywords)
	}
	return nil
}

// -------------------- 处理器 --------------------

// handleMessageCheck 检查消息是否包含屏蔽词
func (bp *BanwordPlugin) handleMessageCheck(ctx *context.Context) {
	// 只处理群组消息
	if ctx.GetChat().Type != "group" && ctx.GetChat().Type != "supergroup" {
		return
	}

	groupID := ctx.GetChat().ID
	message := strings.ToLower(strings.TrimSpace(ctx.GetMessageText()))

	if message == "" {
		return
	}

	bp.db.mu.RLock()
	keywords, exists := bp.db.Groups[groupID]
	bp.db.mu.RUnlock()

	if !exists || len(keywords) == 0 {
		return
	}

	// 检查是否包含屏蔽词
	for _, keyword := range keywords {
		if strings.Contains(message, strings.ToLower(keyword)) {
			// 删除消息
			if err := ctx.DeleteMessage(ctx.GetMessageID()); err != nil {
				bp.Log.Error().Err(err).Msg("删除消息失败")
			} else {
				bp.Log.Info().
					Int64("group_id", groupID).
					Int64("user_id", ctx.GetUserID()).
					Str("keyword", keyword).
					Msg("检测到屏蔽词，已删除消息")
			}

			return
		}
	}
}

// handleAddBanword 添加屏蔽词
func (bp *BanwordPlugin) handleAddBanword(ctx *context.Context, cmdCtx params.CommandContext) {
	// 只允许群组使用
	if ctx.GetChat().Type != "group" && ctx.GetChat().Type != "supergroup" {
		ctx.Reply("❌ 此命令只能在群组中使用")
		return
	}

	if cmdCtx.Args.Len() < 1 {
		ctx.Reply("❌ 用法：添加屏蔽 <关键词1> [关键词2...]")
		return
	}

	groupID := ctx.GetChat().ID

	bp.db.mu.Lock()

	// 获取或创建群组屏蔽词列表
	groupKeywords, exists := bp.db.Groups[groupID]
	if !exists {
		groupKeywords = make([]string, 0)
	}

	// 添加新关键词（去重）
	added := make([]string, 0)
	for i := 0; i < cmdCtx.Args.Len(); i++ {
		kw := strings.TrimSpace(cmdCtx.Args.Get(i))
		if kw == "" {
			continue
		}

		// 检查是否已存在
		exists := false
		for _, existing := range groupKeywords {
			if strings.EqualFold(existing, kw) {
				exists = true
				break
			}
		}

		if !exists {
			groupKeywords = append(groupKeywords, kw)
			added = append(added, kw)
		}
	}

	bp.db.Groups[groupID] = groupKeywords
	bp.db.mu.Unlock()

	// 保存到文件（在锁外执行）
	if err := bp.saveData(); err != nil {
		bp.Log.Error().Err(err).Msg("保存数据失败")
		ctx.Reply("❌ 添加失败")
		return
	}

	if len(added) == 0 {
		ctx.Reply("ℹ️ 所有关键词已存在")
		return
	}

	bp.Log.Info().
		Int64("group_id", groupID).
		Strs("keywords", added).
		Msg("添加屏蔽词成功")

	ctx.Replyf("✅ 添加屏蔽成功\n新增关键词: %s", strings.Join(added, ", "))
}

// handleDeleteBanword 删除屏蔽词
func (bp *BanwordPlugin) handleDeleteBanword(ctx *context.Context, cmdCtx params.CommandContext) {
	// 只允许群组使用
	if ctx.GetChat().Type != "group" && ctx.GetChat().Type != "supergroup" {
		ctx.Reply("❌ 此命令只能在群组中使用")
		return
	}

	if cmdCtx.Args.Len() < 1 {
		ctx.Reply("❌ 用法：删除屏蔽 <关键词1> [关键词2...]")
		return
	}

	groupID := ctx.GetChat().ID

	bp.db.mu.Lock()

	groupKeywords, exists := bp.db.Groups[groupID]
	if !exists || len(groupKeywords) == 0 {
		bp.db.mu.Unlock()
		ctx.Reply("ℹ️ 当前群组没有屏蔽词")
		return
	}

	// 删除关键词
	deleted := make([]string, 0)
	newKeywords := make([]string, 0)

	for _, existing := range groupKeywords {
		shouldDelete := false
		for i := 0; i < cmdCtx.Args.Len(); i++ {
			kw := strings.TrimSpace(cmdCtx.Args.Get(i))
			if strings.EqualFold(existing, kw) {
				shouldDelete = true
				deleted = append(deleted, existing)
				break
			}
		}
		if !shouldDelete {
			newKeywords = append(newKeywords, existing)
		}
	}

	bp.db.Groups[groupID] = newKeywords

	// 如果群组没有屏蔽词了，可以选择删除该条目
	if len(newKeywords) == 0 {
		delete(bp.db.Groups, groupID)
	}

	bp.db.mu.Unlock()

	// 保存到文件（在锁外执行）
	if err := bp.saveData(); err != nil {
		bp.Log.Error().Err(err).Msg("保存数据失败")
		ctx.Reply("❌ 删除失败")
		return
	}

	if len(deleted) == 0 {
		ctx.Reply("ℹ️ 未找到要删除的关键词")
		return
	}

	bp.Log.Info().
		Int64("group_id", groupID).
		Strs("keywords", deleted).
		Msg("删除屏蔽词成功")

	ctx.Replyf("✅ 删除屏蔽成功\n已删除: %s", strings.Join(deleted, ", "))
}

// handleListBanword 查看屏蔽词列表
func (bp *BanwordPlugin) handleListBanword(ctx *context.Context) {
	// 只允许群组使用
	if ctx.GetChat().Type != "group" && ctx.GetChat().Type != "supergroup" {
		ctx.Reply("❌ 此命令只能在群组中使用")
		return
	}

	groupID := ctx.GetChat().ID

	bp.db.mu.RLock()
	keywords, exists := bp.db.Groups[groupID]
	bp.db.mu.RUnlock()

	if !exists || len(keywords) == 0 {
		ctx.Reply("📝 当前群组没有屏蔽词")
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🚫 当前群组屏蔽词列表 (共 %d 个):\n\n", len(keywords)))

	for i, kw := range keywords {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, kw))
	}

	ctx.Reply(sb.String())
}

// -------------------- 数据管理 --------------------

// loadData 从文件加载数据
func (bp *BanwordPlugin) loadData() error {
	// 检查路径是否有效
	if bp.config.DBPath == "" {
		bp.Log.Warn().Msg("数据库路径为空，跳过加载")
		return nil
	}

	data, err := os.ReadFile(bp.config.DBPath)
	if err != nil {
		if os.IsNotExist(err) {
			bp.Log.Info().Msg("数据文件不存在，将使用空数据库")
			return nil // 文件不存在，使用空数据库
		}
		return err
	}

	bp.db.mu.Lock()
	defer bp.db.mu.Unlock()

	if err := json.Unmarshal(data, bp.db); err != nil {
		return err
	}

	return nil
}

// saveData 保存数据到文件
func (bp *BanwordPlugin) saveData() error {
	// 检查路径是否有效
	if bp.config.DBPath == "" {
		return fmt.Errorf("数据库路径为空")
	}

	// 确保目录存在
	dir := filepath.Dir(bp.config.DBPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}

	bp.db.mu.RLock()
	data, err := json.MarshalIndent(bp.db, "", "  ")
	bp.db.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	// 使用临时文件 + 原子重命名，避免写入失败导致数据损坏
	tmpFile := bp.config.DBPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	if err := os.Rename(tmpFile, bp.config.DBPath); err != nil {
		os.Remove(tmpFile) // 清理临时文件
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}
