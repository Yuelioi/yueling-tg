package fortune

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"yueling_tg/core/context"
	"yueling_tg/core/handler"
	"yueling_tg/core/message"
	"yueling_tg/core/on"
	"yueling_tg/core/plugin"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var _ plugin.Plugin = (*FortuneGenerator)(nil)

type FortuneConfig struct {
	BasePath    string
	CacheDir    string
	ThemesDir   string
	FontsDir    string
	Copywriting string
}

type FortuneGenerator struct {
	cfg FortuneConfig
	*plugin.Base
}

func New() plugin.Plugin {

	basePath := "./data/fortune"

	cfg := FortuneConfig{
		BasePath:    basePath,
		CacheDir:    filepath.Join(basePath, "cache"),
		ThemesDir:   filepath.Join(basePath, "themes"),
		FontsDir:    filepath.Join(basePath, "fonts"),
		Copywriting: filepath.Join(basePath, "copywriting.json"),
	}
	p := &FortuneGenerator{
		Base: plugin.NewBase(&plugin.PluginInfo{
			ID:          "fortune",
			Name:        "抽签",
			Description: "发送抽签结果",
			Version:     "0.1.0",
			Author:      "月离",
			Usage:       "抽签 <主题>",
			Extra:       make(map[string]any),
			Group:       "funny",
		}),
		cfg: cfg,
	}

	cmdMatcher := on.OnCommand([]string{"抽签"}, true, handler.NewHandler(p.divine))

	p.AddMatcher(cmdMatcher)
	return p
}

func (fm *FortuneGenerator) divine(ctx *context.Context) {
	uid := ctx.GetUserID()
	// themes := strings.Split(ctx.Args, " ")
	themes := []string{"ba"}
	ok, path, err := fm.Divine(uid, themes...)
	if err != nil {
		ctx.Send(fmt.Sprintf("抽签失败: %s", err.Error()))
		return
	}
	if ok {
		ctx.SendPhoto(message.NewResource(path).WithCaption("✨今日运势✨"))
	} else {
		ctx.SendPhoto(message.NewResource(path).WithCaption("你今天抽过签了，再给你看一次哦🤗"))

	}
}

// ======================
// 抽签入口
// ======================

func (fm *FortuneGenerator) Divine(uid int64, themes ...string) (ok bool, path string, err error) {
	today := time.Now().Format("2006-01-02")

	uidStr := fmt.Sprintf("%d", uid)

	imgPath := filepath.Join(fm.cfg.CacheDir, fmt.Sprintf("%s-%s.png", uidStr, today))

	if _, err := os.Stat(imgPath); err == nil {
		fm.Log.Debug().Str("uid", uidStr).Msg("今天已抽过签，直接返回缓存图片")
		return false, imgPath, nil
	}

	if len(themes) == 0 {
		themes, err = fm.AvailableThemes()
		if err != nil {
			fm.Log.Error().Err(err).Msg("获取主题失败")
			return false, "", err
		}
	}

	theme := themes[rand.Intn(len(themes))]
	fm.Log.Info().Str("uid", uidStr).Str("theme", theme).Msg("开始生成运势图片")

	generatedPath, err := fm.drawing(uidStr, today, theme)
	if err != nil {
		fm.Log.Error().Err(err).Msg("绘制图片失败")
		return false, "", err
	}

	return true, generatedPath, nil
}

// ======================
// 工具函数整合
// ======================

func (fm *FortuneGenerator) AvailableThemes() ([]string, error) {
	themeDirs, err := os.ReadDir(fm.cfg.ThemesDir)
	if err != nil {
		return nil, fmt.Errorf("读取主题目录失败: %w", err)
	}

	var themes []string
	for _, dir := range themeDirs {
		if dir.IsDir() {
			themes = append(themes, dir.Name())
		}
	}
	return themes, nil
}

func (fm *FortuneGenerator) CleanOutPics() error {
	files, err := os.ReadDir(fm.cfg.CacheDir)
	if err != nil {
		return fmt.Errorf("读取缓存目录失败: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			fullPath := filepath.Join(fm.cfg.CacheDir, file.Name())
			if err := os.Remove(fullPath); err != nil {
				fm.Log.Warn().Err(err).Str("file", fullPath).Msg("删除缓存图片失败")
			}
		}
	}
	fm.Log.Info().Msg("缓存清理完成")
	return nil
}

// ======================
// 内部逻辑
// ======================

func (fm *FortuneGenerator) getCopywriting() (string, string, error) {
	data, err := os.ReadFile(fm.cfg.Copywriting)
	if err != nil {
		return "", "", fmt.Errorf("读取文案失败: %w", err)
	}

	var cp struct {
		Copywriting []struct {
			GoodLuck string   `json:"good-luck"`
			Content  []string `json:"content"`
		} `json:"copywriting"`
	}

	if err := json.Unmarshal(data, &cp); err != nil {
		return "", "", fmt.Errorf("解析JSON失败: %w", err)
	}
	if len(cp.Copywriting) == 0 {
		return "", "", fmt.Errorf("文案列表为空")
	}

	item := cp.Copywriting[rand.Intn(len(cp.Copywriting))]
	if len(item.Content) == 0 {
		return "", "", fmt.Errorf("运势(%s)内容为空", item.GoodLuck)
	}
	return item.GoodLuck, item.Content[rand.Intn(len(item.Content))], nil
}

func (fm *FortuneGenerator) loadFont(paths ...string) (*truetype.Font, error) {
	for _, p := range paths {
		fontData, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		font, err := truetype.Parse(fontData)
		if err == nil {
			return font, nil
		}
	}
	return nil, fmt.Errorf("所有字体文件加载失败")
}

func (fm *FortuneGenerator) pickImage(theme string) (string, error) {
	themes, err := fm.AvailableThemes()
	if err != nil {
		return "", err
	}
	if len(themes) == 0 {
		return "", fmt.Errorf("未找到主题")
	}

	picked := theme
	if picked == "" {
		picked = themes[rand.Intn(len(themes))]
	}

	files, err := os.ReadDir(filepath.Join(fm.cfg.ThemesDir, picked))
	if err != nil {
		return "", fmt.Errorf("读取主题(%s)失败: %w", picked, err)
	}
	var imgs []string
	for _, f := range files {
		if !f.IsDir() {
			imgs = append(imgs, filepath.Join(fm.cfg.ThemesDir, picked, f.Name()))
		}
	}
	if len(imgs) == 0 {
		return "", fmt.Errorf("主题(%s)下无图片", picked)
	}
	return imgs[rand.Intn(len(imgs))], nil
}

func (fm *FortuneGenerator) drawing(uid, nowTime, theme string) (string, error) {
	imgPath, err := fm.pickImage(theme)
	if err != nil {
		return "", err
	}

	imgFile, err := os.Open(imgPath)
	if err != nil {
		return "", err
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return "", err
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{}, draw.Src)

	title, text, err := fm.getCopywriting()
	if err != nil {
		return "", err
	}

	font, err := fm.loadFont(filepath.Join(fm.cfg.FontsDir, "sakura.ttf"))
	if err != nil {
		return "", err
	}

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(image.NewUniform(color.White))

	// 标题
	c.SetFontSize(45)
	titleWidth := len([]rune(title)) * 30
	pt := freetype.Pt(135-titleWidth/2, 99+15)
	_, _ = c.DrawString(title, pt)

	// 正文
	c.SetFontSize(25)
	c.SetSrc(image.NewUniform(color.RGBA{50, 50, 50, 255}))

	colNum, columns := decrement(text)
	fontSize := 25
	spacing := 4

	for i := 0; i < colNum; i++ {
		column := columns[i]
		fontHeight := len([]rune(column)) * (fontSize + spacing)
		x := 140 + (colNum-2)*fontSize/2 + (colNum-1)*4 - i*(fontSize+spacing)
		y := 297 - fontHeight/2
		for j, char := range []rune(column) {
			if char == ' ' {
				continue
			}
			pt := freetype.Pt(x, y+j*(fontSize+spacing)+fontSize)
			_, _ = c.DrawString(string(char), pt)
		}
	}

	if err := os.MkdirAll(fm.cfg.CacheDir, 0755); err != nil {
		return "", err
	}

	outPath := filepath.Join(fm.cfg.CacheDir, fmt.Sprintf("%s-%s.png", uid, nowTime))
	outFile, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	if err := png.Encode(outFile, rgba); err != nil {
		return "", err
	}

	fm.Log.Info().Str("output", outPath).Msg("图片生成完成")
	return outPath, nil
}

// ======================
// 文本列排版逻辑
// ======================

func decrement(text string) (int, []string) {
	runes := []rune(text)
	length := len(runes)
	cardinality := 9

	if length > 4*cardinality {
		return 1, []string{text}
	}

	colNum := 1
	tempLen := length
	for tempLen > cardinality {
		colNum++
		tempLen -= cardinality
	}

	if colNum == 2 {
		if length%2 == 0 {
			half := length / 2
			fillIn := strings.Repeat(" ", 9-half)
			return colNum, []string{
				string(runes[:half]) + fillIn,
				fillIn + string(runes[half:]),
			}
		}
		half := (length + 1) / 2
		fillIn := strings.Repeat(" ", 9-half)
		return colNum, []string{
			string(runes[:half]) + fillIn,
			fillIn + " " + string(runes[half:]),
		}
	}

	var result []string
	for i := 0; i < colNum; i++ {
		if i == colNum-1 {
			result = append(result, string(runes[i*cardinality:]))
		} else {
			result = append(result, string(runes[i*cardinality:(i+1)*cardinality]))
		}
	}
	return colNum, result
}
