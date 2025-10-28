package reply

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/handler"
	"yueling_tg/pkg/plugin/params"
)

var _ plugin.Plugin = (*ReplyPlugin)(nil)

// -------------------- 数据结构 --------------------

type ReplyData struct {
	ID      int    `json:"id"`
	QQ      int64  `json:"qq"`      // 添加者 QQ
	Keyword string `json:"keyword"` // 关键词（可以用逗号分隔多个）
	Reply   string `json:"reply"`   // 回复内容
	Group   string `json:"group"`   // 群组（可选）
}

type ReplyDB struct {
	Replies []*ReplyData      `json:"replies"`
	NextID  int               `json:"next_id"`
	Index   map[string]string `json:"-"` // 关键词索引 -> 回复内容（运行时使用）
	mu      sync.RWMutex      `json:"-"`
}

// -------------------- 插件结构 --------------------

type ReplyPlugin struct {
	*plugin.Base
	db     *ReplyDB
	dbPath string
}

func New() plugin.Plugin {
	rp := &ReplyPlugin{
		Base: plugin.NewBase(&plugin.PluginInfo{
			ID:          "reply",
			Name:        "应答系统",
			Description: "基于关键词回复设定内容",
			Version:     "1.0.0",
			Author:      "月离",
			Usage:       "添加回复 <关键词> <回复内容>\n删除回复 <ID>\n更新回复\n查看回复",
			Group:       "系统",
			Extra:       make(map[string]any),
		}),
		dbPath: "./data/reply.json",
		db: &ReplyDB{
			Replies: make([]*ReplyData, 0),
			NextID:  1,
			Index:   make(map[string]string),
		},
	}

	// 加载数据
	if err := rp.loadData(); err != nil {
		rp.Log.Warn().Err(err).Msg("加载回复数据失败，使用空数据库")
	} else {
		rp.Log.Info().Msgf("加载了 %d 条回复数据", len(rp.db.Replies))
	}

	// 更新索引
	rp.updateIndex()

	// 普通消息匹配
	replyHandler := handler.NewHandler(rp.handleReply)
	replyMatcher := plugin.OnMessage(replyHandler).
		SetPriority(1) // 低优先级，让其他命令先处理

	// 添加回复命令
	addHandler := handler.NewHandler(rp.handleAddReply)
	addMatcher := plugin.OnCommand([]string{"添加回复"}, true, addHandler).
		SetPriority(10)

	// 删除回复命令
	deleteHandler := handler.NewHandler(rp.handleDeleteReply)
	deleteMatcher := plugin.OnCommand([]string{"删除回复"}, true, deleteHandler).
		SetPriority(10)

	// 更新回复命令
	updateHandler := handler.NewHandler(rp.handleUpdateReply)
	updateMatcher := plugin.OnCommand([]string{"更新回复"}, true, updateHandler).
		SetPriority(10)

	// 查看回复命令
	listHandler := handler.NewHandler(rp.handleListReply)
	listMatcher := plugin.OnCommand([]string{"查看回复"}, true, listHandler).
		SetPriority(10)

	rp.SetMatchers([]*plugin.Matcher{
		replyMatcher,
		addMatcher,
		deleteMatcher,
		updateMatcher,
		listMatcher,
	})

	return rp
}

// -------------------- 处理器 --------------------

// handleReply 处理普通消息的回复
func (rp *ReplyPlugin) handleReply(ctx *context.Context) {
	msg := strings.ToLower(strings.TrimSpace(ctx.GetMessageText()))
	if msg == "" {
		return
	}

	rp.db.mu.RLock()
	reply, ok := rp.db.Index[msg]
	rp.db.mu.RUnlock()

	if ok {
		// 移除回复前缀（如果有）
		reply = strings.TrimPrefix(reply, "回复")
		if idx := strings.Index(reply, ":"); idx != -1 {
			reply = reply[idx+1:]
		}
		reply = strings.TrimSpace(reply)
		ctx.Reply(reply)
	}
}

// handleAddReply 添加回复
func (rp *ReplyPlugin) handleAddReply(ctx *context.Context, cmdCtx params.CommandContext) {
	if cmdCtx.Args.Len() < 2 {
		ctx.Reply("❌ 用法：添加回复 <关键词> <回复内容>\n支持多关键词：关键词1,关键词2")
		return
	}

	keyword := cmdCtx.Args.Get(0)
	content := strings.Join(cmdCtx.Args[1:], " ")

	// 处理换行符
	content = strings.ReplaceAll(content, "\\n", "\n")

	// 创建新回复
	newReply := &ReplyData{
		ID:      rp.db.NextID,
		QQ:      ctx.GetUserID(),
		Keyword: keyword,
		Reply:   content,
		Group:   "", // 可以根据需要添加群组支持
	}

	rp.db.mu.Lock()
	rp.db.Replies = append(rp.db.Replies, newReply)
	rp.db.NextID++
	rp.db.mu.Unlock()

	// 保存到文件
	if err := rp.saveData(); err != nil {
		rp.Log.Error().Err(err).Msg("保存数据失败")
		ctx.Reply("❌ 添加失败了喵~")
		return
	}

	// 更新索引
	rp.updateIndex()

	rp.Log.Info().
		Int("id", newReply.ID).
		Str("keyword", keyword).
		Msg("添加回复成功")

	ctx.Replyf("✅ 添加成功！回复ID: %d", newReply.ID)
}

// handleDeleteReply 删除回复
func (rp *ReplyPlugin) handleDeleteReply(ctx *context.Context, cmdCtx params.CommandContext) {
	if cmdCtx.Args.Len() < 1 {
		ctx.Reply("❌ 用法：删除回复 <回复ID>")
		return
	}

	idStr := cmdCtx.Args.Get(0)
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		ctx.Reply("❌ 无效的回复ID")
		return
	}

	rp.db.mu.Lock()
	found := false
	newReplies := make([]*ReplyData, 0, len(rp.db.Replies))
	for _, reply := range rp.db.Replies {
		if reply.ID == id {
			found = true
			continue
		}
		newReplies = append(newReplies, reply)
	}
	rp.db.Replies = newReplies
	rp.db.mu.Unlock()

	if !found {
		ctx.Replyf("❌ 未找到ID为 %d 的回复", id)
		return
	}

	// 保存到文件
	if err := rp.saveData(); err != nil {
		rp.Log.Error().Err(err).Msg("保存数据失败")
		ctx.Reply("❌ 删除失败")
		return
	}

	// 更新索引
	rp.updateIndex()

	rp.Log.Info().Int("id", id).Msg("删除回复成功")
	ctx.Reply("✅ 删除成功")
}

// handleUpdateReply 更新索引
func (rp *ReplyPlugin) handleUpdateReply(ctx *context.Context) {
	if err := rp.loadData(); err != nil {
		rp.Log.Error().Err(err).Msg("重新加载数据失败")
		ctx.Reply("❌ 更新失败")
		return
	}

	rp.updateIndex()
	ctx.Replyf("✅ 更新成功！当前共有 %d 条回复", len(rp.db.Replies))
}

// handleListReply 查看回复列表
func (rp *ReplyPlugin) handleListReply(ctx *context.Context) {
	rp.db.mu.RLock()
	defer rp.db.mu.RUnlock()

	if len(rp.db.Replies) == 0 {
		ctx.Reply("📝 当前没有任何回复")
		return
	}

	var sb strings.Builder
	sb.WriteString("📝 回复列表:\n\n")

	for _, reply := range rp.db.Replies {
		content := reply.Reply
		if len(content) > 30 {
			content = content[:30] + "..."
		}
		sb.WriteString(fmt.Sprintf("#%d [%s]\n  → %s\n\n",
			reply.ID,
			reply.Keyword,
			content,
		))
	}

	ctx.Reply(sb.String())
}

// -------------------- 数据管理 --------------------

// updateIndex 更新关键词索引
func (rp *ReplyPlugin) updateIndex() {
	rp.db.mu.Lock()
	defer rp.db.mu.Unlock()

	rp.db.Index = make(map[string]string)

	for _, data := range rp.db.Replies {
		// 分割关键词
		keywords := strings.Split(data.Keyword, ",")
		for _, kw := range keywords {
			kw = strings.TrimSpace(kw)
			if kw == "" {
				continue
			}

			// 处理群组后缀
			suffix := ""
			if data.Group != "" {
				suffix = data.Group
			}

			// 处理 {} 占位符
			reply := data.Reply
			if strings.Contains(reply, "{}") {
				reply = strings.ReplaceAll(reply, "{}", kw)
			}

			// 添加到索引（转小写）
			key := strings.ToLower(kw + suffix)
			rp.db.Index[key] = fmt.Sprintf("回复%d:%s", data.ID, reply)
		}
	}

	rp.Log.Debug().Msgf("更新索引完成，共 %d 个关键词", len(rp.db.Index))
}

// loadData 从文件加载数据
func (rp *ReplyPlugin) loadData() error {
	data, err := os.ReadFile(rp.dbPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，使用空数据库
		}
		return err
	}

	rp.db.mu.Lock()
	defer rp.db.mu.Unlock()

	if err := json.Unmarshal(data, rp.db); err != nil {
		return err
	}

	// 重建索引
	rp.db.Index = make(map[string]string)

	return nil
}

// saveData 保存数据到文件
func (rp *ReplyPlugin) saveData() error {
	// 确保目录存在
	dir := filepath.Dir(rp.dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	rp.db.mu.RLock()
	data, err := json.MarshalIndent(rp.db, "", "  ")
	rp.db.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(rp.dbPath, data, 0644)
}
