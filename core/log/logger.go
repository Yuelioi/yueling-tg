package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Component string

const (
	MiddlewareComponent      Component = "middleware"
	ProviderComponent        Component = "provider"
	BotComponent             Component = "bot"
	APIComponent             Component = "api"
	PluginRegistryComponent  Component = "plugin_registry"
	MatcherRegistryComponent Component = "matcher_registry"
	PluginComponent          Component = "plugin"
	MatcherComponent         Component = "matcher"
	HandlerComponent         Component = "handler"
)

// 创建日志记录器
func New(component Component, name string) zerolog.Logger {
	theme := getComponentTheme(component)

	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",

		FormatLevel: func(i interface{}) string {
			level := strings.ToUpper(fmt.Sprintf("%s", i))
			color := getLevelColor(level)
			return fmt.Sprintf("%s[%-5s]%s", color, level, "\x1b[0m")
		},

		FormatMessage: func(i interface{}) string {
			if i == nil {
				return ""
			}
			// 使用浅蓝色 (Cyan) 显示消息内容
			message := fmt.Sprintf("%s", i)
			return fmt.Sprintf("[%s %s] \x1b[96m%s\x1b[0m", theme.Icon, name, message)
		},

		FormatFieldName: func(i interface{}) string {
			return fmt.Sprintf("\x1b[90m%s\x1b[0m:", i) // 灰色字段名
		},

		FormatFieldValue: func(i interface{}) string {
			fieldStr := fmt.Sprintf("%s", i)
			if strings.Contains(fieldStr, string(component)) {
				return fmt.Sprintf("%s%s%s", theme.Color, fieldStr, "\x1b[0m")
			}
			return fmt.Sprintf("\x1b[97m%s\x1b[0m", fieldStr) // 亮白色
		},

		FormatTimestamp: func(i interface{}) string {
			t, ok := i.(string)
			if !ok {
				return ""
			}
			parsed, err := time.Parse(time.RFC3339, t)
			if err != nil {
				return t
			}
			return fmt.Sprintf("\x1b[90m%s\x1b[0m", parsed.Format("01-02 15:04:05"))
		},
	}

	return zerolog.New(consoleWriter).
		With().
		Timestamp().
		Logger()
}

// 组件主题结构
type ComponentTheme struct {
	Icon  string
	Color string
}

// 根据组件获取主题
func getComponentTheme(component Component) ComponentTheme {
	themes := map[Component]ComponentTheme{
		MiddlewareComponent:     {"🌉", "\x1b[97m"},  // 白色
		ProviderComponent:       {"📦", "\x1b[90m"},  // 灰色
		BotComponent:            {"🚀", "\x1b[94m"},  // 蓝色
		APIComponent:            {"🔌", "\x1b[96m"},  // 青色
		PluginRegistryComponent: {"🎛️", "\x1b[95m"}, // 紫色
		PluginComponent:         {"🧩", "\x1b[93m"},  // 黄色
		MatcherComponent:        {"🔍", "\x1b[35m"},  // 紫红色
		HandlerComponent:        {"⚡", "\x1b[92m"},  // 绿色
	}

	if theme, exists := themes[component]; exists {
		return theme
	}
	return ComponentTheme{"📋", "\x1b[37m"} // 默认主题
}

// getLevelColor 获取日志级别颜色
func getLevelColor(level string) string {
	colors := map[string]string{
		"DEBUG": "\x1b[36m", // 青色
		"INFO":  "\x1b[32m", // 绿色
		"WARN":  "\x1b[33m", // 黄色
		"ERROR": "\x1b[31m", // 红色
		"FATAL": "\x1b[35m", // 紫色
		"PANIC": "\x1b[41m", // 红色背景
	}

	if color, exists := colors[level]; exists {
		return color
	}
	return "\x1b[37m" // 默认白色
}

// NewWithLevel 创建带日志级别的记录器
func NewWithLevel(component Component, name string, level zerolog.Level) zerolog.Logger {
	logger := New(component, name)
	return logger.Level(level)
}

// SetGlobal 设置全局日志记录器
func SetGlobal(component Component, name string) {
	logger := New(component, name)
	zerolog.DefaultContextLogger = &logger
}

func NewMiddleware(name string) zerolog.Logger {
	return New(MiddlewareComponent, name)
}

func NewProvider(name string) zerolog.Logger {
	return New(ProviderComponent, name)
}

func NewBot(name string) zerolog.Logger {
	return New(BotComponent, name)
}

func NewAPI(name string) zerolog.Logger {
	return New(APIComponent, name)
}

func NewPluginRegistry(name string) zerolog.Logger {
	return New(PluginRegistryComponent, name)
}

func NewPlugin(name string) zerolog.Logger {
	return New(PluginComponent, name)
}

func NewMatcher(name string) zerolog.Logger {
	return New(MatcherComponent, name)
}

func NewHandler(name string) zerolog.Logger {
	return New(HandlerComponent, name)
}
