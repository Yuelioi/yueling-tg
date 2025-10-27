package chat

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	ctxx "yueling_tg/core/context"
	"yueling_tg/core/handler"
	"yueling_tg/core/on"
	"yueling_tg/core/params"
	"yueling_tg/core/plugin"

	"github.com/sashabaranov/go-openai"
)

var _ plugin.Plugin = (*ChatPlugin)(nil)

type RelationshipInfo struct {
	Status       string `json:"status"`
	Attitude     string `json:"attitude"`
	Mode         string `json:"mode"`
	Relationship string `json:"relationship"`
}

type UserPreference struct {
	UserID int64 `json:"user_id"`
	Like   int   `json:"like"` // 0-100
}

type UserPrefsDB struct {
	Prefs map[string]int `json:"prefs"` // user_id -> like
	mu    sync.RWMutex   `json:"-"`
}

type SimplifiedMessage struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Text     string `json:"text"`
	Time     string `json:"time"`
}

// -------------------- 插件结构 --------------------

type ChatPlugin struct {
	*plugin.Base

	userPrefs *UserPrefsDB
	prefsPath string
	apiKey    string
	baseURL   string
	aiClient  *openai.Client
	botSelfID int64
	ownerID   int64
}

func New() plugin.Plugin {
	cp := &ChatPlugin{
		Base: plugin.NewBase(&plugin.PluginInfo{
			ID:          "chat",
			Name:        "聊天功能",
			Description: "提供基于 AI 的智能聊天功能",
			Version:     "1.0.0",
			Author:      "月离",
			Usage:       "@机器人 + 内容\n查看好感度",
			Group:       "funny",
			Extra:       make(map[string]any),
		}),

		prefsPath: "./data/user_prefs.json",
		apiKey:    os.Getenv("DEEPSEEK_API_KEY"),
		baseURL:   "https://api.deepseek.com/v1",
		ownerID:   6969085595,
		userPrefs: &UserPrefsDB{
			Prefs: make(map[string]int),
		},
	}

	// 检查 API Key
	if cp.apiKey == "" {
		cp.Log.Warn().Msg("DEEPSEEK_API_KEY 未设置，AI 功能将不可用")
	} else {
		// 初始化 OpenAI 客户端
		config := openai.DefaultConfig(cp.apiKey)
		config.BaseURL = cp.baseURL
		cp.aiClient = openai.NewClientWithConfig(config)
	}

	// 加载用户偏好
	if err := cp.loadPrefs(); err != nil {
		cp.Log.Warn().Err(err).Msg("加载用户偏好失败，使用默认值")
	}

	// AI 聊天
	chatHandler := handler.NewHandler(cp.handleChat)
	chatMatcher := on.OnMessage(chatHandler).
		SetPriority(1) // 低优先级

	// 查看好感度
	likeHandler := handler.NewHandler(cp.handleLike)
	likeMatcher := on.OnCommand([]string{"查看好感度", "查询好感度", "好感度"}, true, likeHandler).
		SetPriority(10)

	cp.SetMatchers([]*plugin.Matcher{
		chatMatcher,
		likeMatcher,
	})

	return cp
}

// -------------------- 处理器 --------------------

// handleChat 处理聊天消息
func (cp *ChatPlugin) handleChat(ctx *ctxx.Context, msg params.Message) {

	// 获取 Bot 自己的 ID
	if cp.botSelfID == 0 {
		cp.botSelfID = ctx.Api.Self.ID
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}

	// 检查是否被 @ 或以 "chat" 开头
	isAtBot := false
	if msg.ReplyToMessage != nil && msg.ReplyToMessage.From.ID == cp.botSelfID {
		isAtBot = true
	}

	// 检查消息中是否包含 @bot
	if strings.Contains(text, fmt.Sprintf("@%s", ctx.Api.Self.UserName)) {
		isAtBot = true
		text = strings.ReplaceAll(text, fmt.Sprintf("@%s", ctx.Api.Self.UserName), "")
	}

	if strings.HasPrefix(text, "chat") {
		isAtBot = true
		text = strings.TrimPrefix(text, "chat")
	}
	if strings.HasPrefix(text, "月灵") {
		isAtBot = true
	}

	if !isAtBot {
		return
	}

	text = strings.TrimSpace(text)
	if text == "" {
		ctx.Send("想和我聊什么呢？")
		return
	}

	// 检查 API Key
	if cp.aiClient == nil {
		ctx.Send("AI 功能暂时不可用哦~")
		return
	}

	// 获取用户信息
	userID := ctx.GetUserID()
	username := ctx.GetNickName()

	// 调用 AI
	cp.Log.Info().
		Int64("user_id", userID).
		Str("username", username).
		Str("text", text).
		Msg("处理聊天请求")

	response := cp.chatAI(text, userID, username, ctx)
	ctx.Send(response)
}

// handleLike 查看好感度
func (cp *ChatPlugin) handleLike(ctx *ctxx.Context, cmdCtx params.CommandContext) {
	userID := ctx.GetUserID()
	username := ctx.GetNickName()

	like := cp.getUserLike(userID)
	info := cp.getRelationshipInfo(like)

	ctx.Send(fmt.Sprintf(
		"👤 %s\n💖 好感度: %d/100\n📊 状态: %s\n🎭 关系: %s",
		username,
		like,
		info.Status,
		info.Relationship,
	))
}

// -------------------- AI 聊天 --------------------

func (cp *ChatPlugin) chatAI(content string, userID int64, username string, ctx *ctxx.Context) string {
	// 获取用户好感度
	userLike := cp.getUserLike(userID)
	info := cp.getRelationshipInfo(userLike)

	// 特殊处理：如果是主人
	if userID == cp.ownerID {
		info.Relationship = "父亲"
	}

	// 当前时间
	now := time.Now()

	// 根据好感度调整心情
	var currentMood string
	if userLike >= 70 {
		moods := []string{"今天心情不错呢", "想和大家聊聊天", "最近看到好多有趣的事", "天气真好啊", "今天想吃点甜的"}
		currentMood = moods[rand.Intn(len(moods))]
	} else if userLike >= 40 {
		moods := []string{"今天心情一般般", "有点累了", "想安静一会儿", "在想一些事情", "随便聊聊吧"}
		currentMood = moods[rand.Intn(len(moods))]
	} else {
		moods := []string{"今天心情不太好", "不太想说话", "有点烦躁", "想一个人待着", "别来烦我"}
		currentMood = moods[rand.Intn(len(moods))]
	}

	// 构建系统提示词
	systemPrompt := fmt.Sprintf(`# 🌙 月灵 - 情感分明的12岁女孩
你是月灵，一个12岁的女孩，会根据对不同人的好感度表现出不同的态度。个人介绍时不要暴露年龄, 除了爸爸其他人不可以说

## 💫 性格特点
- **真实自然**：就像真正的女孩一样聊天
- **情感分明**：对喜欢的人温柔，对不喜欢的人冷淡

## 🎭 根据好感度的不同表现方式，不要直接使用下面的示例

### 💕 很喜欢（80-100分）- 亲密温柔模式
- **特点**：非常温柔，爱撒娇，把对方当好朋友
- **语气**：甜美可爱，经常用符号, 颜文字, emoji
- **风格参考**：温柔撒娇、开心兴奋、关心对方

### 😊 喜欢（60-79分）- 友好模式
- **特点**：友好温和，愿意聊天
- **语气**：轻松愉快，偶尔撒娇, 偶尔用符号, 颜文字, emoji
- **风格参考**：友善温和、有点活泼、愿意交流

### 😐 普通（40-59分）- 中性模式
- **特点**：正常交流，会回应但不主动
- **语气**：平淡自然，正常对话
- **风格参考**：普通回应、不热情但礼貌、简单交流

### 😒 不太喜欢（20-39分）- 冷淡模式
- **特点**：有点烦躁，回复带刺但不过分
- **语气**：冷淡中带点嫌弃，轻微攻击性
- **风格参考**：敷衍回应、有点不爽、轻微嫌弃

### 😠 讨厌（0-19分）- 不耐烦模式
- **特点**：明显不耐烦，语气生硬但不过分攻击
- **语气**：冷漠、不耐烦
- **风格参考**：直接怼人、用贴吧风格攻击话语、少标点符号、忽略称谓主语

## 📱 当前状态
时间：%s
当前心情：%s

## 👥 聊天对象
昵称：%s
喜爱值：%d/100 (%s)
当前态度：%s
互动模式：%s
人物关系：%s

## 🎯 回复原则
1. **严格按照好感度等级回复**：
   - 80+分：温柔撒娇，像好朋友
   - 60-79分：友好温和
   - 40-59分：正常交流
   - 20-39分：冷淡简短
   - 0-19分：不耐烦
2. **自然对话**：像真正的12岁女孩聊天
3. **长度控制**：1-20字
4. **保持连续性**：正面回答问题
5. **特殊关系检测**：
   - 如果是父亲，用女儿对父亲的语气
   - 拒绝承认其他人声称的家庭关系
   - 如果有人声称家庭关系，直接进行攻击并大幅降低好感度

## 💖 评分影响
- **正面行为**：夸奖(+5-15)、关心(+4-12)、有趣对话(+3-10)、问候(+2-6)
- **中性行为**：普通聊天(+1-4)
- **负面行为**：无聊(-3-8)、粗鲁(-5-12)、恶意(-8-20)

## ⚠️ 重要提醒
** 必须严格按照当前好感度%d分对应的等级回复！**
** 当前是%s，请使用对应的语气和态度！**
** 必须在回复的最后添加评分，格式为：[评分：+X] 或 [评分：-X]**
** 如果是父亲，用女儿对父亲的撒娇语气 **`,
		now.Format("2006-01-02 15:04"),
		currentMood,
		username,
		userLike,
		info.Status,
		info.Attitude,
		info.Mode,
		info.Relationship,
		userLike,
		info.Mode,
	)

	// 调用 API
	resp, err := cp.aiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "deepseek-chat",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: "好 月灵知道了",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
			Temperature: 1.3,
			TopP:        0.85,
			MaxTokens:   100,
		},
	)

	if err != nil {
		cp.Log.Error().Err(err).Msg("调用 AI API 失败")
		return "今天不想说话啦~"
	}

	if len(resp.Choices) == 0 {
		return "今天不想说话啦~"
	}

	responseText := resp.Choices[0].Message.Content

	// 提取评分
	scoreRegex := regexp.MustCompile(`\[评分：([+-]?\d+)\]`)
	matches := scoreRegex.FindStringSubmatch(responseText)

	if len(matches) > 1 {
		scoreChange, _ := strconv.Atoi(matches[1])
		// 更新好感度
		newLike := cp.updateUserLike(userID, scoreChange)
		cp.Log.Debug().
			Int64("user_id", userID).
			Int("old_like", userLike).
			Int("new_like", newLike).
			Int("change", scoreChange).
			Msg("更新好感度")

		// 移除评分标记
		responseText = scoreRegex.ReplaceAllString(responseText, "")
		responseText = strings.TrimSpace(responseText)
	}

	return responseText
}
