package sticker

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"yueling_tg/pkg/config"
	"yueling_tg/pkg/plugin"
)

var _ plugin.Plugin = (*StickerPlugin)(nil)

// -------------------- 数据结构 --------------------

// 用户状态
type UserState struct {
	Mode               string // "waiting" - 等待选择贴纸库, "adding" - 正在添加贴纸
	StickerSetID       string // 选中的贴纸库ID (short_name)
	StickerSetTitle    string // 贴纸库标题
	UserID             int64  // 用户ID
	LastUpdate         time.Time
	ProcessCount       int   // 处理次数
	ProcessingMsgID    int   // 正在处理的图片消息ID,
	ProcessingChatID   int64 // 正在处理的聊天ID
	ProcessingInFlight bool  // 是否正在处理图片消息

}

// -------------------- 插件主结构 --------------------

type StickerPlugin struct {
	*plugin.Base
	userStates map[int64]*UserState // userID -> state (用于临时的添加流程状态)
	stateMutex sync.RWMutex
	httpClient *http.Client

	db     *StickerSetDB
	config PluginConfig
}

// 配置结构体
type PluginConfig struct {
	DBPath string `mapstructure:"db_path"` // 数据文件路径
}

func New() plugin.Plugin {
	sp := &StickerPlugin{
		userStates: make(map[int64]*UserState),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		db:         &StickerSetDB{},
	}

	// 插件信息
	info := &plugin.PluginInfo{
		ID:          "sticker",
		Name:        "贴纸管理",
		Description: "管理Telegram贴纸库",
		Version:     "1.0.0",
		Author:      "月离",
		Usage:       "添加贴纸",
		Group:       "工具",
		Extra:       make(map[string]any),
	}

	// 默认配置
	defaultCfg := PluginConfig{
		DBPath: "./data/botsticker.json",
	}

	// 加载或创建配置
	if err := config.GetPluginConfigOrDefault(info.ID, &sp.config, defaultCfg); err != nil {
		panic(fmt.Sprintf("加载插件配置失败: %v", err))
	}

	// 初始化 Builder
	builder := plugin.New().Info(info)

	// 命令注册
	builder.OnCommand("创建贴纸集").Block(true).Do(sp.handleCreateSet)
	builder.OnCommand("删除贴纸集").Do(sp.handleDeleteSet)
	builder.OnCommand("查看贴纸集").Do(sp.handleListSet)

	// 添加贴纸命令
	builder.OnStartsWith("添加贴纸").Do(sp.handleAddSticker)

	// 处理贴纸库选择
	builder.OnCallbackStartsWith(info.ID + ":select:").Priority(9).Do(sp.handleStickerSetSelect)

	// 处理取消
	builder.OnCallbackStartsWith(info.ID + ":cancel").Priority(9).Do(sp.handleCancel)

	// 处理图片消息
	builder.OnMessage().Priority(8).Do(sp.handleImage)

	// 返回插件并注入 Base
	return builder.Go(sp)
}

func (sp *StickerPlugin) Init() error {
	// 加载数据库
	if err := sp.loadData(); err != nil {
		sp.Log.Warn().Msgf("⚠️ 加载贴纸集数据失败，使用空数据库: %v", err)
	} else {
		sp.Log.Info().Msgf("已加载 %d 个贴纸集", len(sp.db.Sets))
	}

	return nil
}
