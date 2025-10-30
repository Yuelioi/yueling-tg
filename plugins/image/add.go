package image

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/common"
	"yueling_tg/pkg/plugin/params"
)

// -------------------- 添加图片逻辑 --------------------
func (rg *RandomGenerator) handleAddImage(c *context.Context, cmdCtx params.CommandContext, commandArgs params.CommandArgs) {
	cmd := cmdCtx.Command
	rg.Log.Info().
		Str("from", c.GetUsername()).
		Str("cmd", string(cmd)).
		Msg("收到添加图片命令")

	categoryKeyword := strings.TrimSpace(strings.TrimPrefix(cmd, "添加"))
	if categoryKeyword == "" {
		c.Reply("请使用正确命令格式，例如：添加老婆 或 添加吃的。")
		return
	}

	// 2. 在配置中查找匹配的分类
	var matchedCategory *CategoryConfig
	for i := range rg.config.Categories {
		cat := &rg.config.Categories[i]
		for _, trigger := range cat.Commands {
			if trigger == categoryKeyword {
				matchedCategory = cat
				break
			}
		}
		if matchedCategory != nil {
			break
		}
	}

	if matchedCategory == nil {
		c.Reply(fmt.Sprintf("未找到对应分类: %s，请检查配置或使用其他命令。", categoryKeyword))
		return
	}

	// 3. 获取 folder
	folderName := matchedCategory.Folder
	if folderName == "" {
		c.Reply("分类配置中未指定文件夹，请检查配置。")
		return
	}

	// 获取图片
	photos, ok := c.GetPhotos()
	if !ok {
		c.Reply("请附带至少一张图片 📷")
		return
	}

	// 构建文件夹路径
	folder := filepath.Join(rg.config.ImagesFolder, folderName)
	if err := os.MkdirAll(folder, 0755); err != nil {
		rg.Log.Error().Err(err).Msg("创建文件夹失败")
		c.Reply("保存图片失败，无法创建文件夹 😢")
		return
	}

	// 拼接命令参数作为文件名前缀
	argParts := []string{}
	for _, a := range commandArgs {
		part := strings.TrimSpace(string(a))
		if part != "" {
			argParts = append(argParts, part)
		}
	}
	namePrefix := strings.Join(argParts, "_")

	success := 0
	duplicates := 0

	for i, fileID := range photos {
		url, err := c.GetFileDirectURL(fileID)
		if err != nil {
			rg.Log.Error().Err(err).Msg("获取文件链接失败")
			c.Replyf("第 %d 张图片获取失败 😭", i+1)
			continue
		}

		data, err := common.FetchFile(url)
		if err != nil {
			rg.Log.Error().Err(err).Msg("下载文件失败")
			c.Replyf("第 %d 张下载失败 😭", i+1)
			continue
		}

		// 计算哈希
		hash := common.Sha1Hash(data)

		// -------------------- 自动识别文件类型 --------------------
		mime := http.DetectContentType(data)
		ext := ".jpg"
		switch {
		case strings.Contains(mime, "png"):
			ext = ".png"
		case strings.Contains(mime, "webp"):
			ext = ".webp"
		case strings.Contains(mime, "gif"):
			ext = ".gif"
		case strings.Contains(mime, "jpeg"):
			ext = ".jpg"
		}

		// -------------------- 生成文件名 --------------------
		var filename string
		if namePrefix != "" {
			// 用户带参数 → 参数名 + 随机4位
			filename = fmt.Sprintf("%s_%s%s", namePrefix, common.RandomString(4), ext)
		} else {
			// 未带参数 → 使用 SHA1 哈希
			filename = fmt.Sprintf("%s%s", hash, ext)
		}

		savePath := filepath.Join(folder, filename)

		// 保存文件
		if err := os.WriteFile(savePath, data, 0644); err != nil {
			rg.Log.Error().Err(err).Msg("保存文件失败")
			c.Replyf("第 %d 张保存失败 😭", i+1)
			continue
		}

		// 添加到索引
		rg.addToIndex(savePath, folderName, filename)

		imgIndex, ok := rg.findHistoryByPath(savePath)
		if !ok || imgIndex == nil {
			imgIndex = &ImageIndex{
				Path:     savePath,
				Filename: filepath.Base(savePath),
				Hash:     hash,
				Category: folderName,
			}

			// 保存历史记录索引
			key := fmt.Sprintf("%d:%d", c.GetChatID().ID, c.GetMessageID())
			rg.msgHistory.Store(key, imgIndex)
		}

		success++
		rg.Log.Info().
			Str("path", savePath).
			Str("hash", hash).
			Msg("图片已保存")
	}

	// 保存索引
	if success > 0 {
		if err := rg.saveIndex(); err != nil {
			rg.Log.Error().Err(err).Msg("保存索引失败")
		}
	}

	// -------------------- 返回结果 --------------------
	var replyMsg string
	if namePrefix != "" {
		if success > 0 {
			replyMsg = fmt.Sprintf("成功保存 %d 张图片为『%s』到『%s』📁", success, namePrefix, folderName)
		} else {
			replyMsg = "没有新图片保存 😢"
		}
	} else {
		if success > 0 && duplicates > 0 {
			replyMsg = fmt.Sprintf("成功保存 %d 张新图片到『%s』📁\n有 %d 张图片已存在，已跳过 🔄", success, folderName, duplicates)
		} else if success > 0 {
			replyMsg = fmt.Sprintf("成功保存 %d 张图片到『%s』📁", success, folderName)
		} else if duplicates > 0 {
			replyMsg = fmt.Sprintf("所有图片都已存在，共 %d 张 ✅", duplicates)
		} else {
			replyMsg = "图片保存失败 😢"
		}
	}

	c.Reply(replyMsg)
}
