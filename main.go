package main

import (
	"net/http"
	"net/url"
	"os"
	"time"
	logx "yueling_tg/internal/core/log"

	"yueling_tg/middleware"
	"yueling_tg/pkg/bot"
	"yueling_tg/plugins/admin"
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

	"github.com/rs/zerolog/log"

	"github.com/joho/godotenv"
)

func main() {

	logger := logx.NewSystem("系统")
	log.Logger = logger
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

	httpProxy := os.Getenv("HTTP_PROXY")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if httpProxy == "" {
		logger.Warn().Msg("未设置代理，将使用默认网络连接")
	} else {

		proxyURL, err := url.Parse(httpProxy)
		if err != nil {
			return
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		client.Transport = transport
		logger.Info().Msgf("已使用环境变量设置代理: HTTP_PROXY=%s", httpProxy)
	}

	b, err := bot.NewBot(botToken, client)
	if err != nil {
		logger.Panic().Msg("创建 Bot 失败")
	}

	b.RegisterMiddlewares(
		middleware.LoggingMiddleware(),
		middleware.RateLimitMiddleware(60, 1*time.Minute),
		middleware.RecoveryMiddleware(),
	)

	b.RegisterPlugins(
		image.New(), emotion.New(), fortune.New(), help.New(), reply.New(), chat.New(),
		ban.New(), recall.New(), calculator.New(), random.New(), music.New(), sticker.New(), admin.New(),
	)

	b.Run()

}
