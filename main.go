package main

import (
	"os"
	"time"
	"yueling_tg/internal/core/log"
	"yueling_tg/middleware"
	"yueling_tg/pkg/bot"
	"yueling_tg/plugins/ban"
	"yueling_tg/plugins/calculator"
	"yueling_tg/plugins/chat"
	"yueling_tg/plugins/emotion"
	"yueling_tg/plugins/fortune"
	"yueling_tg/plugins/help"
	"yueling_tg/plugins/image"
	"yueling_tg/plugins/music"
	"yueling_tg/plugins/random"
	"yueling_tg/plugins/recall"
	"yueling_tg/plugins/reply"
	"yueling_tg/plugins/sticker"

	"github.com/joho/godotenv"
)

func main() {

	logger := log.NewBot("月灵")
	logger.Info().Msg("启动 Telegram Bot...")

	// 加载 .env 文件
	if err := godotenv.Load(".env"); err != nil {
		logger.Warn().Msg("未找到 .env 文件，将使用系统环境变量")
	}

	// 读取 Telegram Bot Token
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		logger.Panic().Msg("TELEGRAM_BOT_TOKEN 未设置")
	}

	// 设置代理
	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")

	if httpProxy == "" || httpsProxy == "" {
		logger.Warn().Msg("未设置代理，将使用默认网络连接")
	} else {
		// 设置代理
		os.Setenv("HTTP_PROXY", httpProxy)
		os.Setenv("HTTPS_PROXY", httpsProxy)
		logger.Info().Msgf("已使用环境变量设置代理: HTTP_PROXY=%s, HTTPS_PROXY=%s", httpProxy, httpsProxy)
	}

	b, err := bot.NewBot(botToken, logger)
	if err != nil {
		logger.Panic().Msg("创建 Bot 失败")
	}

	b.RegisterPlugins(
		image.New(), emotion.New(), fortune.New(), help.New(), reply.New(), chat.New(),
		ban.New(), recall.New(), calculator.New(), random.New(), music.New(), sticker.New())

	b.RegisterMiddlewares(
		middleware.LoggingMiddleware(),
		middleware.RateLimitMiddleware(60, 1*time.Minute),
		middleware.RecoveryMiddleware(),
	)
	b.Run()

}
