package image

import (
	"fmt"
	"os"
	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin/params"
)

// -------------------- 删除图片逻辑 --------------------
func (rg *RandomGenerator) handleDeleteImage(c *context.Context, cmdCtx params.CommandContext, commandArgs params.CommandArgs) {
	replyMsg := c.GetReplyToMessage()
	if replyMsg == nil {
		c.Reply("请回复一条图片消息进行删除")
		return
	}

	key := fmt.Sprintf("%d:%d", replyMsg.Chat.ID, replyMsg.MessageID)
	val, ok := rg.msgHistory.Load(key)
	if !ok || val == nil {
		c.Reply("未找到历史记录，可能不是通过本系统发送的图片/已被删除")
		return
	}

	imgIndex, ok := val.(*ImageIndex)
	if !ok || imgIndex == nil {
		c.Reply("图片索引数据无效")
		return
	}

	if err := os.Remove(imgIndex.Path); err != nil {
		rg.Log.Error().Err(err).Str("path", imgIndex.Path).Msg("删除文件失败")
		c.Reply("删除文件失败 😭")
		return
	}

	rg.removeFromIndex(imgIndex.Hash)
	rg.msgHistory.Delete(key)
	rg.saveIndex()

	c.Replyf("已删除 %s 分类的图片：%s ✅", imgIndex.Category, imgIndex.Filename)
}
