package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
	logx "yueling_tg/internal/core/log"

	"github.com/rs/zerolog/log"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
)

type ZerologWrapper struct{}

// Debugf logs debug messages
func (z ZerologWrapper) Debugf(format string, args ...any) {
	log.Debug().Msgf(format, args...)
}

// Errorf logs error messages
func (z ZerologWrapper) Errorf(format string, args ...any) {
	log.Error().Msgf(format, args...)
}

func main() {

	logger3 := logx.NewAPI("API")
	log.Logger = logger3

	logger := logx.NewBot("月灵")
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
	// httpProxy := os.Getenv("HTTP_PROXY")
	// httpsProxy := os.Getenv("HTTPS_PROXY")

	// if httpProxy == "" || httpsProxy == "" {
	// 	logger.Warn().Msg("未设置代理，将使用默认网络连接")
	// } else {
	// 	// 设置代理
	// 	os.Setenv("HTTP_PROXY", httpProxy)
	// 	os.Setenv("HTTPS_PROXY", httpsProxy)
	// 	logger.Info().Msgf("已使用环境变量设置代理: HTTP_PROXY=%s, HTTPS_PROXY=%s", httpProxy, httpsProxy)
	// }

	logger2 := ZerologWrapper{}

	proxyURL, err := url.Parse("http://127.0.0.1:10808") // HTTP 代理
	if err != nil {
		return
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	// ---- SOCKS5 代理 ----
	// dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:1080", nil, proxy.Direct)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// transport := &http.Transport{
	//     Dial: dialer.Dial,
	// }

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	// Note: Please keep in mind that default logger may expose sensitive information,
	// use in development only
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger(), telego.WithHTTPClient(client), telego.WithLogger(logger2))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Get updates channel
	// (more on configuration in examples/updates_long_polling/main.go)
	updates, _ := bot.UpdatesViaLongPolling(context.Background(), nil)

	// Loop through all updates when they came
	for update := range updates {

		fmt.Printf("Update: %+v\n", update)
	}
}
