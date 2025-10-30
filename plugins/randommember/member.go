package randommember

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"

	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/config"
	"yueling_tg/pkg/plugin"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

var _ plugin.Plugin = (*RandomMemberPlugin)(nil)

// -------------------- 数据结构 --------------------

// MemberInfo 成员信息
type MemberInfo struct {
	UserID       int64     `json:"user_id"`
	Username     string    `json:"username"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	IsBot        bool      `json:"is_bot"`
	LastActivity time.Time `json:"last_activity"`
}

// GroupMembers 群组成员数据
type GroupMembers struct {
	Members map[int64]map[int64]*MemberInfo `json:"members"` // chatID -> userID -> MemberInfo
	mu      sync.RWMutex                    `json:"-"`
}

// -------------------- 插件结构 --------------------

type PluginConfig struct {
	DBPath      string `mapstructure:"db_path"`
	MaxMembers  int    `mapstructure:"max_members"`  // 每个群最多保留多少活跃成员
	ActiveLimit int    `mapstructure:"active_limit"` // 从最近多少活跃成员中抽取
	AllowBots   bool   `mapstructure:"allow_bots"`   // 是否允许抽到机器人
}

type RandomMemberPlugin struct {
	*plugin.Base
	data   *GroupMembers
	config PluginConfig
}

func New() plugin.Plugin {
	rmp := &RandomMemberPlugin{
		data: &GroupMembers{
			Members: make(map[int64]map[int64]*MemberInfo),
		},
	}

	// 插件信息
	info := &plugin.PluginInfo{
		ID:          "random_member",
		Name:        "随机群友",
		Description: "随机抽取群友（从最近活跃成员中）",
		Version:     "1.0.0",
		Author:      "月离",
		Usage:       "抽群友 / 来点群友 / 随机群友",
		Group:       "随机",
		Extra:       make(map[string]any),
	}

	// 设置默认配置
	rmp.config = PluginConfig{
		DBPath:      "./data/random_member.json",
		MaxMembers:  100,  // 每个群最多保留100个活跃成员
		ActiveLimit: 25,   // 从最近25个活跃成员中抽取
		AllowBots:   true, // 允许抽到机器人
	}

	// 尝试加载配置
	if err := config.GetPluginConfigOrDefault(info.ID, &rmp.config, rmp.config); err != nil {
		rmp.config.DBPath = "./data/random_member.json"
		rmp.config.MaxMembers = 100
		rmp.config.ActiveLimit = 25
		rmp.config.AllowBots = true
	}

	// 确保路径不为空
	if rmp.config.DBPath == "" {
		rmp.config.DBPath = "./data/random_member.json"
	}

	// 初始化 Builder
	builder := plugin.New().Info(info)

	// 追踪所有消息（用于记录活跃成员）
	builder.OnMessage().Priority(1).Do(rmp.trackMember)

	// 注册正则匹配
	builder.OnRegex(`抽(.*)群友(.*)|随机.*群友.*|来个.*群友.*|来点.*群友.*`).
		Priority(5).
		Do(rmp.handleRandomMember)

	// 返回插件，并注入 Base
	return builder.Go(rmp)
}

func (rmp *RandomMemberPlugin) Init() error {
	// 加载数据
	if err := rmp.loadData(); err != nil {
		rmp.Log.Warn().Msgf("⚠️ 加载成员数据失败，使用空数据库: %v", err)
	} else {
		totalMembers := 0
		for _, members := range rmp.data.Members {
			totalMembers += len(members)
		}
		rmp.Log.Info().Msgf("已加载 %d 个群组的活跃成员，共 %d 人", len(rmp.data.Members), totalMembers)
	}
	return nil
}

// -------------------- 处理器 --------------------

// trackMember 追踪活跃成员
func (rmp *RandomMemberPlugin) trackMember(ctx *context.Context) {
	// 只追踪群组消息
	if ctx.GetChat().Type != "group" && ctx.GetChat().Type != "supergroup" {
		return
	}

	chatID := ctx.GetChat().ID
	user := ctx.GetUser()

	rmp.data.mu.Lock()

	// 确保群组存在
	if rmp.data.Members[chatID] == nil {
		rmp.data.Members[chatID] = make(map[int64]*MemberInfo)
	}

	// 更新或添加成员信息
	rmp.data.Members[chatID][user.ID] = &MemberInfo{
		UserID:       user.ID,
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		IsBot:        user.IsBot,
		LastActivity: time.Now(),
	}

	// 限制成员数量（保留最近活跃的）
	if len(rmp.data.Members[chatID]) > rmp.config.MaxMembers {
		rmp.cleanupOldMembers(chatID)
	}

	rmp.data.mu.Unlock()

	go func() {
		// 随机延迟，避免频繁保存
		if rand.Intn(10) == 0 { // 10% 概率保存
			if err := rmp.saveData(); err != nil {
				rmp.Log.Error().Err(err).Msg("保存成员数据失败")
			}
		}
	}()
}

// cleanupOldMembers 清理不活跃的成员（需在锁内调用）
func (rmp *RandomMemberPlugin) cleanupOldMembers(chatID int64) {
	members := rmp.data.Members[chatID]
	if len(members) <= rmp.config.MaxMembers {
		return
	}

	// 转换为切片并按活跃时间排序
	type memberPair struct {
		userID int64
		info   *MemberInfo
	}

	pairs := make([]memberPair, 0, len(members))
	for userID, info := range members {
		pairs = append(pairs, memberPair{userID, info})
	}

	// 按活跃时间降序排序（冒泡排序）
	for i := 0; i < len(pairs)-1; i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[i].info.LastActivity.Before(pairs[j].info.LastActivity) {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	// 只保留最近活跃的
	newMembers := make(map[int64]*MemberInfo)
	for i := 0; i < rmp.config.MaxMembers && i < len(pairs); i++ {
		newMembers[pairs[i].userID] = pairs[i].info
	}

	rmp.data.Members[chatID] = newMembers
	rmp.Log.Debug().Msgf("群组 %d 清理旧成员，保留 %d 人", chatID, len(newMembers))
}

// handleRandomMember 处理随机抽群友
func (rmp *RandomMemberPlugin) handleRandomMember(ctx *context.Context) {
	// 只在群组中工作
	if ctx.GetChat().Type != "group" && ctx.GetChat().Type != "supergroup" {
		ctx.Reply("❌ 此命令只能在群组中使用")
		return
	}

	chatID := ctx.GetChat().ID

	rmp.data.mu.RLock()
	members, exists := rmp.data.Members[chatID]
	rmp.data.mu.RUnlock()

	if !exists || len(members) == 0 {
		ctx.Reply("❌ 还没有活跃成员记录，让大家多聊聊天吧~")
		return
	}

	// 转换为切片并按活跃时间排序
	memberList := make([]*MemberInfo, 0, len(members))
	for _, info := range members {
		// 根据配置决定是否包含机器人
		if !rmp.config.AllowBots && info.IsBot {
			continue
		}
		memberList = append(memberList, info)
	}

	if len(memberList) == 0 {
		ctx.Reply("❌ 没有符合条件的群友")
		return
	}

	// 按活跃时间降序排序（最近的在前）
	for i := 0; i < len(memberList)-1; i++ {
		for j := i + 1; j < len(memberList); j++ {
			if memberList[i].LastActivity.Before(memberList[j].LastActivity) {
				memberList[i], memberList[j] = memberList[j], memberList[i]
			}
		}
	}

	// 只从最近活跃的成员中选择
	limit := rmp.config.ActiveLimit
	if len(memberList) < limit {
		limit = len(memberList)
	}
	activeMembers := memberList[:limit]

	// 随机选择一个成员
	rand.Seed(time.Now().UnixNano())
	selected := activeMembers[rand.Intn(len(activeMembers))]

	// 构建名称
	name := selected.FirstName
	if selected.LastName != "" {
		name += " " + selected.LastName
	}
	if selected.Username != "" {
		name = "@" + selected.Username
	}

	// 添加机器人标识
	botTag := ""
	if selected.IsBot {
		botTag = " 🤖"
	}

	// 获取用户头像
	photos, err := ctx.Api.GetUserProfilePhotos(ctx.Ctx, &telego.GetUserProfilePhotosParams{
		UserID: selected.UserID,
		Limit:  1,
	})

	var photoFileID string
	if err == nil && photos.TotalCount > 0 && len(photos.Photos) > 0 && len(photos.Photos[0]) > 0 {
		// 获取最大尺寸的头像
		photoFileID = photos.Photos[0][len(photos.Photos[0])-1].FileID
	}

	// 发送消息
	text := fmt.Sprintf("🎲 你抽到的群友是: %s%s",
		name,
		botTag,
	)

	if photoFileID != "" {
		// 发送带头像的消息
		_, err = ctx.Api.SendPhoto(ctx.Ctx, &telego.SendPhotoParams{
			ChatID:  tu.ID(chatID),
			Photo:   tu.FileFromID(photoFileID),
			Caption: text,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: ctx.GetMessageID(),
			},
		})
		if err != nil {
			rmp.Log.Error().Err(err).Msg("发送头像失败")
			// 降级为纯文本
			ctx.Reply(text)
		}
	} else {
		// 没有头像，发送纯文本
		ctx.Reply(text)
	}

	rmp.Log.Info().
		Int64("chat_id", chatID).
		Int64("selected_user_id", selected.UserID).
		Str("selected_user_name", name).
		Bool("is_bot", selected.IsBot).
		Msg("随机抽取群友成功")
}

// -------------------- 数据管理 --------------------

// loadData 从文件加载数据
func (rmp *RandomMemberPlugin) loadData() error {
	if rmp.config.DBPath == "" {
		rmp.Log.Warn().Msg("数据库路径为空，跳过加载")
		return nil
	}

	data, err := os.ReadFile(rmp.config.DBPath)
	if err != nil {
		if os.IsNotExist(err) {
			rmp.Log.Info().Msg("数据文件不存在，将使用空数据库")
			return nil
		}
		return err
	}

	rmp.data.mu.Lock()
	defer rmp.data.mu.Unlock()

	if err := json.Unmarshal(data, rmp.data); err != nil {
		return err
	}

	return nil
}

// saveData 保存数据到文件
func (rmp *RandomMemberPlugin) saveData() error {
	if rmp.config.DBPath == "" {
		return fmt.Errorf("数据库路径为空")
	}

	// 确保目录存在
	dir := filepath.Dir(rmp.config.DBPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}

	rmp.data.mu.RLock()
	data, err := json.MarshalIndent(rmp.data, "", "  ")
	rmp.data.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	// 使用临时文件 + 原子重命名
	tmpFile := rmp.config.DBPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	if err := os.Rename(tmpFile, rmp.config.DBPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}

// -------------------- 辅助函数 --------------------

// formatDuration 格式化时间间隔
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "刚刚"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d分钟前", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1天前"
	}
	if days < 30 {
		return fmt.Sprintf("%d天前", days)
	}
	months := days / 30
	if months == 1 {
		return "1个月前"
	}
	if months < 12 {
		return fmt.Sprintf("%d个月前", months)
	}
	years := months / 12
	return fmt.Sprintf("%d年前", years)
}
