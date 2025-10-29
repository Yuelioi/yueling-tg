package random

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"yueling_tg/internal/core/context"
	"yueling_tg/pkg/plugin"
	"yueling_tg/pkg/plugin/handler"
	"yueling_tg/pkg/plugin/params"
)

var _ plugin.Plugin = (*RollPlugin)(nil)

func New() plugin.Plugin {
	rp := &RollPlugin{
		Base: plugin.NewBase(&plugin.PluginInfo{
			ID:          "roll",
			Name:        "随心Roll",
			Description: "支持数字随机、字符串随机、GIF/视频随机帧抽取(需要ffmpeg)",
			Version:     "1.1.0",
			Author:      "月离",
			Usage:       "roll 整数 | 整数 整数 | x y z... | GIF/视频",
			Group:       "娱乐",
			Extra:       make(map[string]any),
		}),
	}

	cmds := []string{"roll"}
	handler := handler.NewHandler(rp.rollHandler)
	m := plugin.OnCommand(cmds, true, handler)
	rp.AddMatcher(m)

	return rp
}

type RollPlugin struct {
	*plugin.Base
}

// ----------------------------------------------
// Roll 核心逻辑
// ----------------------------------------------

func (rp *RollPlugin) rollHandler(c *context.Context, commandArgs params.CommandArgs) {
	rp.Log.Info().Msgf("Roll 指令参数: %v", commandArgs)

	// 	// 处理图片/视频
	// if photos, ok := c.GetMedias(); ok {
	// 	fileID := photos[0]
	// 	rp.handleMedia(c, fileID)
	// 	return
	// }

	// 处理命令参数
	var args []string
	for _, arg := range commandArgs {
		args = append(args, strings.TrimSpace(string(arg)))
	}

	if len(args) == 0 {
		c.Reply("用法：roll 整数 | 整数 整数 | x y z... | GIF/视频")
		return
	}

	if len(args) == 1 && args[0] == "6" {
		c.SendDice("🎲")
		return
	}

	// 情况 1：两个数字 => roll 3 6
	if len(args) == 2 && isInteger(args[0]) && isInteger(args[1]) {
		count, _ := strconv.Atoi(args[0])
		sides, _ := strconv.Atoi(args[1])
		if count > 30 {
			count = 30
		}
		nums := uniqueRandomNumbers(1, sides, count)
		c.Reply(fmt.Sprintf("你 roll 到的数组是 %v 🎲", nums))
		return
	}

	// 情况 2：单个数字 => roll 10
	if len(args) == 1 && isInteger(args[0]) {
		max, _ := strconv.Atoi(args[0])
		if max <= 0 {
			c.Reply("数字要大于 0 哦～")
			return
		}
		num := randInt(1, max)
		c.Reply(fmt.Sprintf("您 roll 到的数字是「%d」 🎯", num))
		return
	}

	botName := c.GetBotFullname()

	// 情况 3：字符串随机 => roll 猫 狗 鸟
	if len(args) >= 2 {
		choices := args
		replyList := []string{
			"emmm，要不试试「%s」？",
			"来试试「%s」吧～",
			botName + "觉得「%s」不错哟~",
			"就决定是你了！「%s」！",
			botName + "想要「%s」！",
		}
		pick := choices[rand.Intn(len(choices))]
		c.Reply(fmt.Sprintf(replyList[rand.Intn(len(replyList))], pick))
		return
	}

	c.Reply("请输入正确的指令 (整数 | 整数 整数 | x y z...)")
}

// ----------------------------------------------
// 辅助函数
// ----------------------------------------------

func isInteger(s string) bool {
	matched, _ := regexp.MatchString(`^-?\d+$`, s)
	return matched
}

func randInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func uniqueRandomNumbers(min, max, count int) []int {
	nums := make([]int, 0, count)
	used := make(map[int]bool)
	for len(nums) < count {
		n := randInt(min, max)
		if !used[n] {
			used[n] = true
			nums = append(nums, n)
		}
	}
	return nums
}
