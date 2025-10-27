package jm

import "encoding/json"

// Response structures
type JmResp struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}

type SearchRespData struct {
	Total       json.Number  `json:"total"`
	Content     []ComicBrief `json:"content"`
	CurrentPage int          `json:"currentPage"`
}

type RedirectRespData struct {
	RedirectAid string `json:"redirect_aid"`
}

type ComicBrief struct {
	Id          json.Number `json:"id"`
	Title       string      `json:"name"`
	Author      []string    `json:"author"`
	Description string      `json:"description"`
	Tags        []string    `json:"tags"`
}

type GetComicRespData struct {
	Id          json.Number    `json:"id"`
	Title       string         `json:"name"`
	Author      []string       `json:"author"`
	Description string         `json:"description"`
	Tags        []string       `json:"tags"`
	Series      []ChapterBrief `json:"series"`
	// 添加其他需要的字段
}

type ChapterBrief struct {
	Id    json.Number `json:"id"`
	Title string      `json:"name"`
	Sort  json.Number `json:"sort"`
}

type GetChapterRespData struct {
	Id       json.Number `json:"id"`
	Title    string      `json:"name"`
	Images   []string    `json:"images"`
	DataSize int         `json:"data_size"`
	// 添加其他需要的字段
}

type WeeklyCategory struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type GetWeeklyInfoRespData struct {
	Categories []WeeklyCategory `json:"categories"`
}

type GetWeeklyRespData struct {
	Comics []ComicBrief `json:"content"`
}

// ImageInfo 包含图片URL和对应的block_num
type ImageInfo struct {
	URL      string
	BlockNum uint32
	Index    int
}
