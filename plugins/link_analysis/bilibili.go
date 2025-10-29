package link_analysis

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"yueling_tg/internal/core/context"
	"yueling_tg/internal/message"

	"github.com/valyala/fasthttp"

	"yueling_tg/pkg/plugin"
)

// -------------------- 插件结构 --------------------

var _ plugin.Plugin = (*BiliPlugin)(nil)

type BiliPlugin struct {
	*plugin.Base
	lastLink string
}

func NewBili() plugin.Plugin {
	bp := &BiliPlugin{}

	info := &plugin.PluginInfo{
		ID:          "bili",
		Name:        "B站链接解析",
		Description: "解析消息中的 B站 视频/番剧/直播/专栏/动态 链接并返回信息",
		Version:     "1.0.0",
		Author:      "月离",
		Usage:       "发送 B站 链接即可解析",
		Group:       "工具",
		Extra:       make(map[string]any),
	}

	builder := plugin.New().Info(info)

	// 注册正则匹配命令
	builder.OnRegex(`https?://[^\s]+`).Priority(1).Block(true).Do(bp.handleBiliLink)

	// 返回插件，并注入 Base
	return builder.Go(bp)
}

// -------------------- 处理器 --------------------

func (bp *BiliPlugin) handleBiliLink(ctx *context.Context) {
	text := ctx.GetMessageText()
	if text == "" {
		return
	}

	if text == bp.lastLink {
		return
	}
	bp.lastLink = text

	segments, url := bp.ParseMessage(ctx)
	if url == "" {
		return
	}

	for _, seg := range segments {
		if seg.Type == "text" {
			ctx.Send(seg.Data)
		} else if seg.Type == "image" {
			ctx.SendPhoto(message.NewResource(seg.Data))
		}
	}
}

type MessageSegment struct {
	Type string // "text" 或 "image"
	Data string
}

func (bp *BiliPlugin) ParseMessage(ctx *context.Context) ([]MessageSegment, string) {
	text := ctx.GetMessageText()
	if text == "" {
		return nil, ""
	}

	urlStr, page, tParam := bp.extract(text)
	if urlStr == "" {
		return nil, ""
	}

	switch {
	case strings.Contains(urlStr, "view?"):
		return bp.videoDetail(urlStr, page, tParam)
	case strings.Contains(urlStr, "bangumi"):
		return bp.bangumiDetail(urlStr, tParam)
	case strings.Contains(urlStr, "xlive"):
		return bp.liveDetail(urlStr)
	case strings.Contains(urlStr, "/read/"):
		return bp.articleDetail(urlStr, page)
	case strings.Contains(urlStr, "dynamic"):
		return bp.dynamicDetail(urlStr)
	}

	return nil, ""
}

func (bp *BiliPlugin) fetchJSON(url string, target interface{}) error {
	status, body, err := fasthttp.Get(nil, url)
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("HTTP %d", status)
	}
	return json.Unmarshal(body, target)
}

func handleNum(num int) string {
	if num >= 10000 {
		return fmt.Sprintf("%.2f万", float64(num)/10000)
	}
	return fmt.Sprintf("%d", num)
}

func (bp *BiliPlugin) extract(text string) (string, string, string) {
	page := ""
	tParam := ""

	if m := regexp.MustCompile(`[?&]p=(\d+)`).FindStringSubmatch(text); len(m) > 1 {
		page = m[1]
	}
	if m := regexp.MustCompile(`[?&]t=(\d+)`).FindStringSubmatch(text); len(m) > 1 {
		tParam = m[1]
	}

	if m := regexp.MustCompile(`av(\d+)`).FindStringSubmatch(text); len(m) > 1 {
		return fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?aid=%s", m[1]), page, tParam
	}
	if m := regexp.MustCompile(`BV([A-Za-z0-9]{10})`).FindStringSubmatch(text); len(m) > 1 {
		return fmt.Sprintf("https://api.bilibili.com/x/web-interface/view?bvid=%s", m[0]), page, tParam
	}
	if m := regexp.MustCompile(`ep(\d+)`).FindStringSubmatch(text); len(m) > 1 {
		return fmt.Sprintf("https://api.bilibili.com/pgc/view/web/season?ep_id=%s", m[1]), page, tParam
	}
	if m := regexp.MustCompile(`ss(\d+)`).FindStringSubmatch(text); len(m) > 1 {
		return fmt.Sprintf("https://api.bilibili.com/pgc/view/web/season?season_id=%s", m[1]), page, tParam
	}
	if m := regexp.MustCompile(`md(\d+)`).FindStringSubmatch(text); len(m) > 1 {
		return fmt.Sprintf("https://api.bilibili.com/pgc/review/user?media_id=%s", m[1]), page, tParam
	}
	if m := regexp.MustCompile(`live.bilibili.com/(?:blanc/|h5/)?(\d+)`).FindStringSubmatch(text); len(m) > 1 {
		return fmt.Sprintf("https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom?room_id=%s", m[1]), page, tParam
	}
	if m := regexp.MustCompile(`/read/(?:cv|mobile|native)/?(\d+)`).FindStringSubmatch(text); len(m) > 1 {
		return fmt.Sprintf("https://api.bilibili.com/x/article/viewinfo?id=%s&mobi_app=pc&from=web", m[1]), m[1], tParam
	}
	if m := regexp.MustCompile(`bilibili.com.*?/(\d+)\?.*?type=2`).FindStringSubmatch(text); len(m) > 1 {
		return fmt.Sprintf("https://api.vc.bilibili.com/dynamic_svr/v1/dynamic_svr/get_dynamic_detail?rid=%s&type=2", m[1]), page, tParam
	}
	if m := regexp.MustCompile(`bilibili.com.*?/(\d+)`).FindStringSubmatch(text); len(m) > 1 {
		return fmt.Sprintf("https://api.vc.bilibili.com/dynamic_svr/v1/dynamic_svr/get_dynamic_detail?dynamic_id=%s", m[1]), page, tParam
	}

	return "", page, tParam
}

// ------------------- 视频 -------------------
func (bp *BiliPlugin) videoDetail(urlStr, page, tParam string) ([]MessageSegment, string) {
	var res struct {
		Data struct {
			Aid   int    `json:"aid"`
			Bvid  string `json:"bvid"`
			Title string `json:"title"`
			Pic   string `json:"pic"`
			Owner struct {
				Name string `json:"name"`
			} `json:"owner"`
			Stat struct {
				View, Danmaku, Favorite, Like, Coin, Reply int
			} `json:"stat"`
			Desc    string `json:"desc"`
			Tname   string `json:"tname"`
			Pubdate int64  `json:"pubdate"`
		} `json:"data"`
	}

	if err := bp.fetchJSON(urlStr, &res); err != nil {
		return []MessageSegment{{Type: "text", Data: fmt.Sprintf("解析视频失败: %v", err)}}, ""
	}

	vurl := fmt.Sprintf("https://www.bilibili.com/video/av%d", res.Data.Aid)
	pubdate := time.Unix(res.Data.Pubdate, 0).Format("2006-01-02 15:04:05")

	msgs := []MessageSegment{
		{Type: "image", Data: res.Data.Pic},
		{Type: "text", Data: fmt.Sprintf("%s\n类型：%s | UP：%s | 日期：%s\n播放：%s | 弹幕：%s | 收藏：%s | 点赞：%s | 硬币：%s | 评论：%s\n简介：%s",
			vurl, res.Data.Tname, res.Data.Owner.Name, pubdate,
			handleNum(res.Data.Stat.View), handleNum(res.Data.Stat.Danmaku),
			handleNum(res.Data.Stat.Favorite), handleNum(res.Data.Stat.Like),
			handleNum(res.Data.Stat.Coin), handleNum(res.Data.Stat.Reply),
			res.Data.Desc)},
	}

	return msgs, vurl
}

// ------------------- 番剧 -------------------
func (bp *BiliPlugin) bangumiDetail(urlStr, tParam string) ([]MessageSegment, string) {
	var res struct {
		Result struct {
			Cover    string   `json:"cover"`
			Title    string   `json:"title"`
			Style    []string `json:"style"`
			Episodes []struct {
				EpID       int    `json:"ep_id"`
				IndexTitle string `json:"index_title"`
			} `json:"episodes"`
			NewestEp struct {
				Desc string `json:"desc"`
			} `json:"newest_ep"`
			Evaluate string `json:"evaluate"`
			SeasonID int    `json:"season_id"`
			MediaID  int    `json:"media_id"`
		} `json:"result"`
	}

	if err := bp.fetchJSON(urlStr, &res); err != nil {
		return []MessageSegment{{Type: "text", Data: fmt.Sprintf("解析番剧失败: %v", err)}}, ""
	}

	style := strings.Join(res.Result.Style, ", ")
	msgs := []MessageSegment{
		{Type: "image", Data: res.Result.Cover},
		{Type: "text", Data: fmt.Sprintf("番剧：%s\n类型：%s\n简介：%s\n最新一集：%s",
			res.Result.Title, style, res.Result.Evaluate, res.Result.NewestEp.Desc)},
	}

	vurl := fmt.Sprintf("https://www.bilibili.com/bangumi/play/ss%d", res.Result.SeasonID)
	return msgs, vurl
}

// ------------------- 直播 -------------------
func (bp *BiliPlugin) liveDetail(urlStr string) ([]MessageSegment, string) {
	var res struct {
		Code int `json:"code"`
		Data struct {
			RoomInfo struct {
				Title          string `json:"title"`
				RoomID         int    `json:"room_id"`
				LiveStatus     int    `json:"live_status"`
				LockStatus     bool   `json:"lock_status"`
				ParentAreaName string `json:"parent_area_name"`
				AreaName       string `json:"area_name"`
				Online         int    `json:"online"`
				Cover          string `json:"cover"`
				Tags           string `json:"tags"`
			} `json:"room_info"`
			AnchorInfo struct {
				BaseInfo struct {
					Uname string `json:"uname"`
				} `json:"base_info"`
			} `json:"anchor_info"`
			WatchedShow struct {
				TextLarge string `json:"text_large"`
			} `json:"watched_show"`
		} `json:"data"`
	}

	if err := bp.fetchJSON(urlStr, &res); err != nil || res.Code != 0 {
		return []MessageSegment{{Type: "text", Data: fmt.Sprintf("解析直播失败: %v", err)}}, ""
	}

	title := res.Data.RoomInfo.Title
	status := "[未开播]"
	if res.Data.RoomInfo.LiveStatus == 1 {
		status = "[直播中]"
	} else if res.Data.RoomInfo.LiveStatus == 2 {
		status = "[轮播中]"
	}
	msgs := []MessageSegment{
		{Type: "image", Data: res.Data.RoomInfo.Cover},
		{Type: "text", Data: fmt.Sprintf("%s %s\n主播：%s 当前分区：%s-%s\n在线人数：%d\n标签：%s",
			status, title, res.Data.AnchorInfo.BaseInfo.Uname,
			res.Data.RoomInfo.ParentAreaName, res.Data.RoomInfo.AreaName,
			res.Data.RoomInfo.Online, res.Data.RoomInfo.Tags)},
	}

	vurl := fmt.Sprintf("https://live.bilibili.com/%d", res.Data.RoomInfo.RoomID)
	return msgs, vurl
}

// ------------------- 专栏 -------------------
func (bp *BiliPlugin) articleDetail(urlStr, cvid string) ([]MessageSegment, string) {
	var res struct {
		Data struct {
			Title           string   `json:"title"`
			AuthorName      string   `json:"author_name"`
			Mid             int      `json:"mid"`
			OriginImageURLs []string `json:"origin_image_urls"`
			Stats           struct {
				View, Favorite, Coin, Share, Like, Dislike int
			} `json:"stats"`
		} `json:"data"`
	}

	if err := bp.fetchJSON(urlStr, &res); err != nil {
		return []MessageSegment{{Type: "text", Data: fmt.Sprintf("解析专栏失败: %v", err)}}, ""
	}

	msgs := []MessageSegment{}
	for _, img := range res.Data.OriginImageURLs {
		msgs = append(msgs, MessageSegment{Type: "image", Data: img})
	}

	msgs = append(msgs, MessageSegment{Type: "text", Data: fmt.Sprintf(
		"标题：%s\n作者：%s (https://space.bilibili.com/%d)\n阅读：%s 收藏：%s 硬币：%s 分享：%s 点赞：%s 不喜欢：%s",
		res.Data.Title, res.Data.AuthorName, res.Data.Mid,
		handleNum(res.Data.Stats.View), handleNum(res.Data.Stats.Favorite), handleNum(res.Data.Stats.Coin),
		handleNum(res.Data.Stats.Share), handleNum(res.Data.Stats.Like), handleNum(res.Data.Stats.Dislike),
	)})

	vurl := fmt.Sprintf("https://www.bilibili.com/read/cv%s", cvid)
	return msgs, vurl
}

// ------------------- 动态 -------------------
func (bp *BiliPlugin) dynamicDetail(urlStr string) ([]MessageSegment, string) {
	var res struct {
		Data struct {
			Card      string `json:"card"`
			DynamicID string `json:"dynamic_id"`
		} `json:"data"`
	}

	if err := bp.fetchJSON(urlStr, &res); err != nil {
		return []MessageSegment{{Type: "text", Data: fmt.Sprintf("解析动态失败: %v", err)}}, ""
	}

	var card map[string]interface{}
	if err := json.Unmarshal([]byte(res.Data.Card), &card); err != nil {
		return []MessageSegment{{Type: "text", Data: "动态 JSON 解析失败"}}, ""
	}

	content := ""
	if item, ok := card["item"].(map[string]interface{}); ok {
		if desc, ok2 := item["description"].(string); ok2 {
			content = desc
		} else if desc2, ok3 := item["content"].(string); ok3 {
			content = desc2
		}
	}

	msgs := []MessageSegment{{Type: "text", Data: content}}
	vurl := fmt.Sprintf("https://t.bilibili.com/%s", res.Data.DynamicID)
	return msgs, vurl
}
