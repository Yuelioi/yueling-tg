package music

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"

	"github.com/mymmrac/telego"
)

var _ plugin.Plugin = (*MusicPlugin)(nil)

// -------------------- 数据结构 --------------------

type SearchResult struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Artist  []string `json:"artist"`
	Album   string   `json:"album"`
	PicID   string   `json:"pic_id"`
	LyricID string   `json:"lyric_id"`
	Source  string   `json:"source"`
}

type URLResult struct {
	URL  string `json:"url"`
	BR   int    `json:"br"`
	Size int    `json:"size"`
}

// 音乐源配置
var musicSources = []struct {
	ID   string
	Name string
}{
	{"netease", "网易云"},
	// {"kuwo", "酷我"},
	{"joox", "JOOX"},
	// {"tencent", "QQ音乐"},
	// {"kugou", "酷狗"},
	// {"migu", "咪咕"},
}

// 搜索缓存结构
type SearchCache struct {
	Results []SearchResult
	Keyword string
	Source  string
	Time    time.Time
}

// -------------------- 插件主结构 --------------------

type MusicPlugin struct {
	*plugin.Base
	apiBase     string
	httpClient  *http.Client
	searchCache map[int64]*SearchCache // chatID -> cache
	cacheMutex  sync.RWMutex

	limit int // 最多显示多少结果
}

func New() plugin.Plugin {
	mp := &MusicPlugin{
		Base: plugin.NewBase(&plugin.PluginInfo{
			ID:          "music",
			Name:        "点歌插件",
			Description: "支持多平台搜索点歌",
			Version:     "2.0.0",
			Author:      "月离",
			Usage:       "点歌 <歌曲名>",
			Group:       "娱乐",
		}),
		apiBase:     "https://music-api.gdstudio.xyz/api.php",
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		searchCache: make(map[int64]*SearchCache),
		limit:       10,
	}

	builder := plugin.New().
		Info(mp.PluginInfo())

	// 点歌命令
	builder.OnStartsWith("点歌").
		Do(mp.handleSearch)

	// 处理音乐源切换
	builder.OnCallbackStartsWith(mp.PluginInfo().ID + ":source:").
		Priority(9).
		Do(mp.handleSourceChange)

	// 处理歌曲播放
	builder.OnCallbackStartsWith(mp.PluginInfo().ID + ":play:").
		Priority(9).
		Do(mp.handlePlay)

	return builder.Go()
}

// -------------------- 搜索处理 --------------------

func (mp *MusicPlugin) limitResult(results []SearchResult) []SearchResult {
	if len(results) < mp.limit {
		return results
	}
	return results[:mp.limit]
}

func (mp *MusicPlugin) handleSearch(c *context.Context) {
	text := strings.TrimSpace(c.GetMessageText())
	if text == "" {
		return
	}

	parts := strings.Fields(text)
	if len(parts) < 2 {
		c.Reply("请输入要搜索的歌曲名，例如：点歌 告白气球")
		return
	}

	keyword := strings.Join(parts[1:], " ")

	// 默认使用网易云搜索
	results, err := mp.searchMusic("netease", keyword, 5)
	if err != nil {
		c.Reply(fmt.Sprintf("搜索失败：%v", err))
		return
	}

	if len(results) == 0 {
		c.Reply("没有找到相关歌曲 😢")
		return
	}

	results = mp.limitResult(results)

	// 缓存搜索结果
	chatID := c.GetChatID()
	mp.cacheMutex.Lock()
	mp.searchCache[chatID.ID] = &SearchCache{
		Results: results,
		Keyword: keyword,
		Source:  "netease",
		Time:    time.Now(),
	}
	mp.cacheMutex.Unlock()

	mp.showSearchResults(c, results, keyword, "netease")
}

// -------------------- 音乐源切换 --------------------

func (mp *MusicPlugin) handleSourceChange(cmd string, c *context.Context) error {
	// 格式: music:source:SOURCE
	parts := strings.Split(cmd, ":")
	if len(parts) < 3 {
		c.AnswerCallback("参数错误")
		return nil
	}

	source := parts[2]

	// 从缓存获取关键词
	chatID := c.GetChatID()
	mp.cacheMutex.RLock()
	cache, exists := mp.searchCache[chatID.ID]
	mp.cacheMutex.RUnlock()

	if !exists {
		c.AnswerCallback("搜索已过期，请重新搜索")
		return nil
	}

	keyword := cache.Keyword

	results, err := mp.searchMusic(source, keyword, 5)
	if err != nil {
		c.AnswerCallback(fmt.Sprintf("搜索失败：%v", err))
		return nil
	}

	if len(results) == 0 {
		c.AnswerCallback("该音乐源没有找到相关歌曲")
		return nil
	}
	results = mp.limitResult(results)

	// 更新缓存
	mp.cacheMutex.Lock()
	mp.searchCache[chatID.ID] = &SearchCache{
		Results: results,
		Keyword: keyword,
		Source:  source,
		Time:    time.Now(),
	}
	mp.cacheMutex.Unlock()

	msg := c.GetCallbackQuery().Message.Message()
	if msg == nil {
		return nil
	}

	// 更新消息显示新的搜索结果
	mp.updateSearchResults(c, msg, results, keyword, source)
	c.AnswerCallback("已切换音乐源")
	return nil
}

// -------------------- 播放处理 --------------------

// -------------------- 播放处理 --------------------

func (mp *MusicPlugin) handlePlay(cmd string, c *context.Context) error {
	// 格式: music:play:INDEX
	parts := strings.Split(cmd, ":")
	if len(parts) < 3 {
		c.AnswerCallback("参数错误")
		return nil
	}

	var index int
	fmt.Sscanf(parts[2], "%d", &index)

	// 从缓存获取搜索结果
	chatID := c.GetChatID()
	mp.cacheMutex.RLock()
	cache, exists := mp.searchCache[chatID.ID]
	mp.cacheMutex.RUnlock()

	if !exists || index < 0 || index >= len(cache.Results) {
		c.AnswerCallback("数据已过期，请重新搜索")
		return nil
	}

	song := cache.Results[index]
	trackID := song.ID
	songName := song.Name
	artist := strings.Join(song.Artist, ", ")
	source := song.Source

	mp.Log.Debug().Msgf("Playing: source=%s, id=%s, name=%s", source, trackID, songName)

	// 获取音乐URL
	urlResult, err := mp.getMusicURL(source, trackID)
	if err != nil || urlResult.URL == "" {
		c.AnswerCallback("获取音乐链接失败 😢")
		return nil
	}

	msg := c.GetCallbackQuery().Message
	if msg == nil {
		return nil
	}

	// 发送音频消息
	params := &telego.SendAudioParams{
		ChatID:    c.GetChatID(),
		Audio:     telego.InputFile{FileID: urlResult.URL}, // 如果是URL，使用FileID字段
		Title:     songName,
		Performer: artist,
		ReplyParameters: &telego.ReplyParameters{
			MessageID: msg.GetMessageID(),
		},
	}

	// 如果urlResult.URL是HTTP URL，需要用不同方式
	if strings.HasPrefix(urlResult.URL, "http") {
		params.Audio = telego.InputFile{URL: urlResult.URL}
	}

	_, err = c.Api.SendAudio(c.Ctx, params)
	if err != nil {
		// 如果发送音频失败，发送链接
		sizeStr := fmt.Sprintf("%.2f MB", float64(urlResult.Size)/1024)
		linkMsg := fmt.Sprintf("🎵 %s - %s\n\n🔗 播放链接：\n%s\n\n音质：%dkbps | 大小：%s",
			songName, artist, urlResult.URL, urlResult.BR, sizeStr)

		replyParams := &telego.SendMessageParams{
			ChatID: c.GetChatID(),
			Text:   linkMsg,
			ReplyParameters: &telego.ReplyParameters{
				MessageID: msg.GetMessageID(),
			},
		}

		c.Api.SendMessage(c.Ctx, replyParams)
		c.AnswerCallback("已发送播放链接")
	} else {
		c.AnswerCallback("正在播放 🎶")
	}

	return nil
}

// -------------------- 显示搜索结果 --------------------

func (mp *MusicPlugin) showSearchResults(c *context.Context, results []SearchResult, keyword, source string) {
	var buttons [][]telego.InlineKeyboardButton

	// 添加歌曲按钮
	for i, song := range results {
		artist := strings.Join(song.Artist, ", ")
		buttonText := fmt.Sprintf("%d. %s - %s", i+1, song.Name, artist)
		if len(buttonText) > 60 {
			buttonText = buttonText[:57] + "..."
		}

		// 使用索引而不是ID
		callbackData := fmt.Sprintf("%s:play:%d", mp.PluginInfo().ID, i)

		buttons = append(buttons, []telego.InlineKeyboardButton{
			{
				Text:         buttonText,
				CallbackData: callbackData,
			},
		})
	}

	// 添加音乐源切换按钮
	var sourceButtons []telego.InlineKeyboardButton
	for _, src := range musicSources {
		emoji := ""
		if src.ID == source {
			emoji = "✓ "
		}
		callbackData := fmt.Sprintf("%s:source:%s", mp.PluginInfo().ID, src.ID)

		sourceButtons = append(sourceButtons, telego.InlineKeyboardButton{
			Text:         emoji + src.Name,
			CallbackData: callbackData,
		})

		// 每行3个按钮
		if len(sourceButtons) == 3 {
			buttons = append(buttons, sourceButtons)
			sourceButtons = []telego.InlineKeyboardButton{}
		}
	}
	if len(sourceButtons) > 0 {
		buttons = append(buttons, sourceButtons)
	}

	markup := telego.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	msgText := fmt.Sprintf("🔍 搜索结果：%s\n📱 请选择要播放的歌曲：", keyword)
	c.SendMessageWithMarkup(msgText, markup)
}

// -------------------- 更新搜索结果 --------------------

func (mp *MusicPlugin) updateSearchResults(c *context.Context, msg *telego.Message, results []SearchResult, keyword, source string) {
	var buttons [][]telego.InlineKeyboardButton

	// 添加歌曲按钮
	for i, song := range results {
		artist := strings.Join(song.Artist, ", ")
		buttonText := fmt.Sprintf("%d. %s - %s", i+1, song.Name, artist)
		if len(buttonText) > 60 {
			buttonText = buttonText[:57] + "..."
		}

		// 使用索引而不是ID
		callbackData := fmt.Sprintf("%s:play:%d", mp.PluginInfo().ID, i)

		buttons = append(buttons, []telego.InlineKeyboardButton{
			{
				Text:         buttonText,
				CallbackData: callbackData,
			},
		})
	}

	// 添加音乐源切换按钮
	var sourceButtons []telego.InlineKeyboardButton
	for _, src := range musicSources {
		emoji := ""
		if src.ID == source {
			emoji = "✓ "
		}
		callbackData := fmt.Sprintf("%s:source:%s", mp.PluginInfo().ID, src.ID)

		sourceButtons = append(sourceButtons, telego.InlineKeyboardButton{
			Text:         emoji + src.Name,
			CallbackData: callbackData,
		})

		if len(sourceButtons) == 3 {
			buttons = append(buttons, sourceButtons)
			sourceButtons = []telego.InlineKeyboardButton{}
		}
	}
	if len(sourceButtons) > 0 {
		buttons = append(buttons, sourceButtons)
	}

	markup := telego.InlineKeyboardMarkup{
		InlineKeyboard: buttons,
	}

	sourceName := ""
	for _, src := range musicSources {
		if src.ID == source {
			sourceName = src.Name
			break
		}
	}

	msgText := fmt.Sprintf("🔍 搜索结果：%s\n📱 当前音乐源：%s\n\n请选择要播放的歌曲：", keyword, sourceName)

	params := &telego.EditMessageTextParams{
		ChatID:      c.GetChatID(),
		MessageID:   msg.GetMessageID(),
		Text:        msgText,
		ReplyMarkup: &markup,
	}

	c.Api.EditMessageText(c.Ctx, params)
}

// -------------------- API调用 --------------------

func (mp *MusicPlugin) searchMusic(source, keyword string, count int) ([]SearchResult, error) {
	apiURL := fmt.Sprintf("%s?types=search&source=%s&name=%s&count=%d",
		mp.apiBase, source, url.QueryEscape(keyword), count)

	resp, err := mp.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 先解析成 map 来处理混合类型的 ID
	var rawResults []map[string]interface{}
	if err := json.Unmarshal(body, &rawResults); err != nil {
		return nil, err
	}

	// 转换为 SearchResult，确保 ID 是字符串
	results := make([]SearchResult, 0, len(rawResults))
	for _, raw := range rawResults {
		result := SearchResult{
			Name:   getStringValue(raw, "name"),
			Album:  getStringValue(raw, "album"),
			Source: getStringValue(raw, "source"),
		}

		// 处理 ID - 可能是字符串或数字
		result.ID = convertToString(raw["id"])
		result.PicID = convertToString(raw["pic_id"])
		result.LyricID = convertToString(raw["lyric_id"])

		// 处理 Artist 数组
		if artists, ok := raw["artist"].([]interface{}); ok {
			result.Artist = make([]string, 0, len(artists))
			for _, a := range artists {
				if artistStr, ok := a.(string); ok {
					result.Artist = append(result.Artist, artistStr)
				}
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// 辅助函数：将 interface{} 转换为字符串
func convertToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		// 避免科学计数法
		return fmt.Sprintf("%.0f", val)
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// 辅助函数：从 map 中获取字符串值
func getStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return ""
}

func (mp *MusicPlugin) getMusicURL(source, trackID string) (*URLResult, error) {
	apiURL := fmt.Sprintf("%s?types=url&source=%s&id=%s&br=320",
		mp.apiBase, source, trackID)

	mp.Log.Debug().Msg(apiURL)

	resp, err := mp.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result URLResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
