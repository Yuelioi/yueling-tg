package image

import (
	"os"
	"yueling_tg/core/context"
	"yueling_tg/core/params"
	"yueling_tg/core/utils"
)

// -------------------- 删除图片逻辑 --------------------
func (rg *RandomGenerator) handleDeleteImage(c *context.Context, cmdCtx params.CommandContext, commandArgs params.CommandArgs) {
	rg.Log.Info().
		Str("from", c.GetUsername()).
		Msg("收到删除图片命令")

	replyMsg := c.GetReplyToMessage()

	// 必须回复一张图片
	if replyMsg == nil || len(replyMsg.Photo) == 0 {
		c.Reply("请回复一张图片来删除 🖼️\n使用方法：回复要删除的图片，然后发送「删除图片」命令")
		return
	}

	// 获取最大尺寸的图片
	photo := replyMsg.Photo[len(replyMsg.Photo)-1]
	url, err := c.Api.GetFileDirectURL(photo.FileID)
	if err != nil {
		rg.Log.Error().Err(err).Msg("获取文件链接失败")
		c.Reply("获取图片失败 😭")
		return
	}

	// 下载图片并计算哈希
	data, err := utils.FetchFile(url)
	if err != nil {
		rg.Log.Error().Err(err).Msg("下载文件失败")
		c.Reply("下载图片失败 😭")
		return
	}

	hash := utils.Sha1Hash(data)

	// 在索引中查找
	imgIndex, ok := rg.findByHash(hash)
	if !ok {
		c.Reply("未找到该图片的记录，可能不是通过本系统添加的 🤔")
		return
	}

	// 删除文件
	if err := os.Remove(imgIndex.Path); err != nil {
		rg.Log.Error().Err(err).Str("path", imgIndex.Path).Msg("删除文件失败")
		c.Reply("删除文件失败 😭")
		return
	}

	// 从索引中移除
	rg.removeFromIndex(hash)

	// 保存索引
	if err := rg.saveIndex(); err != nil {
		rg.Log.Error().Err(err).Msg("保存索引失败")
	}

	rg.Log.Info().
		Str("hash", hash).
		Str("path", imgIndex.Path).
		Str("category", imgIndex.Category).
		Msg("图片已删除")

	c.Replyf("已成功删除『%s』分类的图片：%s ✅", imgIndex.Category, imgIndex.Filename)
}
