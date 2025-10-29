package sticker

import (
	"net/http"
	"sync"
	"time"
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
	dbPath string
}

func New() plugin.Plugin {
	sp := &StickerPlugin{
		Base: plugin.NewBase(&plugin.PluginInfo{
			ID:          "sticker",
			Name:        "贴纸管理",
			Description: "管理Telegram贴纸库",
			Version:     "1.0.0",
			Author:      "月离",
			Usage:       "添加贴纸",
			Group:       "工具",
		}),
		userStates: make(map[int64]*UserState),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		db:         &StickerSetDB{},
		dbPath:     "./data/botsticker.json",
	}

	// 加载数据
	if err := sp.loadData(); err != nil {
		sp.Log.Warn().Err(err).Msg("加载贴纸集数据失败，使用空数据库")
	} else {
		sp.Log.Info().Msgf("加载了 %d 个贴纸集", len(sp.db.Sets))
	}

	builder := plugin.New().
		Info(sp.PluginInfo())

	builder.OnCommand("创建贴纸集").Block(true).Do(sp.handleCreateSet)
	builder.OnCommand("删除贴纸集").Do(sp.handleDeleteSet)
	builder.OnCommand("查看贴纸集").Do(sp.handleListSet)

	// 添加贴纸命令
	builder.OnStartsWith("添加贴纸").
		Do(sp.handleAddSticker)

	// 处理贴纸库选择
	builder.OnCallbackStartsWith(sp.PluginInfo().ID + ":select:").
		Priority(9).
		Do(sp.handleStickerSetSelect)

	// 处理取消
	builder.OnCallbackStartsWith(sp.PluginInfo().ID + ":cancel").
		Priority(9).
		Do(sp.handleCancel)

	// 处理图片消息
	builder.OnMessage().
		Priority(8).
		Do(sp.handleImage)

	return builder.Go()
}
