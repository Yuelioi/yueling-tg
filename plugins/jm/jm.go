package jm

import (
	"fmt"
	_ "image/png"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"yueling_tg/internal/core/context"
	"yueling_tg/internal/message"
	"yueling_tg/pkg/config"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/params"

	"github.com/jung-kurt/gofpdf"
)

var _ plugin.Plugin = (*JMPlugin)(nil)

var (
	BaseURL    = "https://18comic.vip/"
	ProxyURL   = "http://127.0.0.1:10808" // 代理地址
	SaveDir    = "./data/jm"
	MaxWorkers = 5
	RetryTimes = 3
	WaitTime   = time.Second
)

type JMPlugin struct {
	*plugin.Base
	client *JmClient
	config PluginConfig
}

type PluginConfig struct {
	ProxyURL string `json:"proxy_url"`
	SaveDir  string `json:"save_dir"`
}

func New() *JMPlugin {
	jm := &JMPlugin{}

	// 插件信息
	info := &plugin.PluginInfo{
		ID:          "jm",
		Name:        "JM 下载器",
		Description: "下载 JM 漫画并生成 PDF",
		Version:     "1.0.0",
		Author:      "月离",
		Usage:       "jm <书籍ID> [章节号]",
		Group:       "娱乐",
		Extra:       make(map[string]any),
	}

	// 默认配置
	defaultCfg := PluginConfig{
		ProxyURL: "http://127.0.0.1:7890",
		SaveDir:  "./data/downloads",
	}

	// 加载或创建配置
	if err := config.GetPluginConfigOrDefault(info.ID, &jm.config, defaultCfg); err != nil {
		panic(fmt.Sprintf("加载插件配置失败: %v", err))
	}

	// 初始化 Builder
	builder := plugin.New().Info(info)

	// 注册命令
	builder.OnCommand("jm").Priority(1).Do(jm.handleJM)

	// 创建带代理的 HTTP 客户端
	transport := &http.Transport{
		Proxy: http.ProxyURL(MustParseURL(jm.config.ProxyURL)),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	jm.client = NewJmClient(client).
		WithConcurrency(8).
		WithSaveDir(jm.config.SaveDir)

	// 返回插件并注入 Base
	return builder.Go(jm).(*JMPlugin)
}

// -------------------- 主处理函数 --------------------

func (jm *JMPlugin) handleJM(c *context.Context, cmdCtx params.CommandContext) {
	args := cmdCtx.Args

	if len(args) < 1 {
		c.Reply("请输入正确的命令格式: /jm <书籍ID> [章节号]")
		return
	}

	bookID, err := strconv.Atoi(args[0])
	if err != nil {
		c.Reply("请输入正确的书籍ID")
		return
	}
	chapterIndex := 1

	if len(args) >= 2 {
		if num, err := strconv.Atoi(args[1]); err == nil {
			chapterIndex = abs(num)
		}
	}

	jm.Log.Info().
		Int("bookID", bookID).
		Int("chapter index", chapterIndex).
		Msg("开始处理 JM 下载")

	// 创建保存目录
	tmpFolder := filepath.Join(SaveDir, args[0])
	if err := os.MkdirAll(tmpFolder, 0755); err != nil {
		jm.Log.Error().Err(err).Msg("创建目录失败")
		c.Reply("创建目录失败")
		return
	}

	// 解析章节
	jm.Log.Info().Msg("正在解析章节...")
	comic, err := jm.client.GetComic(bookID)
	if err != nil || len(comic.Series) == 0 {
		jm.Log.Error().Err(err).Msg("解析章节失败")
		c.Reply("网络错误(请稍后重试)/未找到任何章节")
		return
	}

	if chapterIndex > len(comic.Series) {
		c.Reply(fmt.Sprintf("最多只有 %d 章节喔", len(comic.Series)))
		return
	}

	chapterID := comic.Series[chapterIndex-1].Id

	pdfFile := filepath.Join(tmpFolder, fmt.Sprintf("%d_%d.pdf", bookID, chapterIndex))

	if _, err := os.Stat(pdfFile); err == nil {
		jm.Log.Info().Str("pdf文件路径", pdfFile).Msg("PDF 文件已存在")
		c.SendDocument(message.NewResource(pdfFile))
		return
	}

	c.Send(fmt.Sprintf("下载%s中, 请稍后 %d/%d", comic.Title, chapterIndex, len(comic.Series)))

	cid, _ := chapterID.Int64()

	// 下载图片
	if err := jm.client.DownloadChapterImages(int(cid), ImageFormatPNG); err != nil {
		jm.Log.Error().Err(err).Msg("下载图片失败")
		c.Reply(fmt.Sprintf("下载出错: %v", err))
		return
	}

	// 生成 PDF
	jm.Log.Info().Msg("正在生成 PDF...")
	if err := jm.convertImagesToPDF(tmpFolder, pdfFile); err != nil {
		jm.Log.Error().Err(err).Msg("生成 PDF 失败")
		c.Reply("生成 PDF 失败")
		return
	}

	c.SendDocument(message.NewResource(pdfFile))
}

// -------------------- 生成 PDF --------------------

func (jm *JMPlugin) convertImagesToPDF(imageFolder, outputPDF string) error {
	files, err := os.ReadDir(imageFolder)
	if err != nil {
		return err
	}

	var imagePaths []string
	validExts := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".webp": true,
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if validExts[ext] {
			imagePaths = append(imagePaths, filepath.Join(imageFolder, f.Name()))
		}
	}

	if len(imagePaths) == 0 {
		return fmt.Errorf("没有找到图片文件")
	}

	// 排序（按文件名自然顺序）
	sort.Strings(imagePaths)

	// 创建 PDF（A4 纵向）
	pdf := gofpdf.New("P", "mm", "A4", "")
	pageW, pageH := pdf.GetPageSize()

	for _, imgPath := range imagePaths {
		pdf.AddPage()

		// 注册图片
		opt := gofpdf.ImageOptions{ImageType: "", ReadDpi: true}
		info := pdf.RegisterImageOptions(imgPath, opt)

		imgW, imgH := info.Extent()
		ratio := min(pageW/imgW, pageH/imgH)

		newW := imgW * ratio
		newH := imgH * ratio

		// 居中绘制
		x := (pageW - newW) / 2
		y := (pageH - newH) / 2

		pdf.ImageOptions(imgPath, x, y, newW, newH, false, opt, 0, "")
	}

	if err := pdf.OutputFileAndClose(outputPDF); err != nil {
		return fmt.Errorf("写入 PDF 失败: %w", err)
	}

	jm.Log.Info().Str("output", outputPDF).Msg("PDF 生成成功")
	return nil
}

func MustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
