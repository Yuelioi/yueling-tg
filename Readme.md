# Yueling Telegram Bot 框架

一个基于 Go 的 Telegram 聊天机器人框架，支持插件化开发、命令匹配器、权限控制、条件规则以及消息/回调事件处理。框架内置多种示例插件，包括随机图片、表情包、骰子和点歌功能。

---

## 🚀 特性

* **插件化架构**：支持 DSL 风格注册插件和命令。
* **中间件链**：支持自定义中间件处理消息。
* **规则/条件系统**：可灵活组合条件（All / Any / Not）。
* **权限系统**：插件可定义权限规则。
* **消息类型支持**：文本消息、回调按钮、通知、媒体消息。

## 📁 项目结构

```
yueling_tg/
├── internal/
│   ├── core/              // 核心逻辑
│   │   ├── context/       // 上下文与消息封装
│   │   ├── log/           // 日志模块
│   │   └── plugin/        // 内部插件相关逻辑
│   ├── message/           // 消息处理工具（图片、文本等）
│   └── middleware/        // 中间件链
└── pkg/
    ├── bot/               // Bot 核心逻辑
    ├── common/            // 公共工具函数
    └── plugin/            // 插件系统
        ├── dsl/           // DSL 构建器
        │   ├── condition/ // 条件规则
        │   ├── permission/ // 权限模块
        │   └── rule/       // 规则封装
        ├── handler/       // 处理器封装
        ├── params/        // 参数解析
        └── provider/      // 数据/上下文提供者
└── main.go                 // 启动入口
```

---

## ⚙️ 安装与运行

```
# 克隆项目
git clone https://github.com/yuelioiyueling_tg.git
cd yueling_tg
# 下载依赖
go mod tidy
# 运行 Bot
go run main.go
```

> 需在 `.env` 中设置 Telegram Bot Token。

---

## 🔌 插件开发指南

框架采用 **Builder DSL**，插件开发流程：

1. **创建插件结构体**

```
type MyPlugin struct {
    *plugin.Base
}
```

1. **实现插件入口**

```
func NewMyPlugin() plugin.Plugin {
    mp := &MyPlugin{
        Base: plugin.NewBase(&plugin.PluginInfo{
            ID:          "myplugin",
            Name:        "示例插件",
            Description: "这是一个示例插件",
            Version:     "0.1.0",
            Author:      "月离",
            Group:       "demo",
        }),
    }
    builder := plugin.New().
        Info(mp.PluginInfo())
    // 注册消息匹配器
    builder.OnMessage().Do(mp.handleMessage)
    // 注册命令匹配器
    builder.OnCommand("示例").Do(mp.handleCommand)
    return builder.Go()
}
```

1. **处理消息 / 命令**

```
func (mp *MyPlugin) handleMessage(c *context.Context) {
    c.Reply("收到消息啦!")
}
func (mp *MyPlugin) handleCommand(c *context.Context) {
    c.Reply("命令已触发!")
}
```

1. **再来一张 / 再来一首**

* 使用回调按钮 + `OnCallbackStartsWith`
* 修改原消息实现按钮功能

---

## 📦 示例插件

### 随机图片插件

* 支持发送随机图片
* 支持回调按钮 “再来一张”

### 表情包插件

* 支持 `#关键词` 随机表情包
* 支持查询表情包列表 `##关键词`
* 支持 “再来一张” 回调按钮

### 骰子插件

* 支持投掷 🎲 / 🎯 / 🏀 / ⚽ / 🧭
* 命令：`骰子 [类型]`

### 点歌插件

* 命令：`点歌 <歌曲名>`
* 回调按钮：`再来一首`
* 随机更换歌曲并更新原消息

## todo

配置中心

---

## ⚡ 使用 Builder 构建插件

| 方法                                | 说明                 |
| :---------------------------------- | :------------------- |
| OnMessage()                         | 匹配所有文本消息     |
| OnCommand(cmds ...string)           | 匹配指定命令         |
| OnCallbackStartsWith(prefix string) | 匹配回调事件         |
| Priority(int)                       | 设置匹配器优先级     |
| Block(bool)                         | 是否阻止事件继续传播 |
| Do(handlerFn)                       | 绑定处理函数         |

---

## 📝 示例：注册插件

```

bot.RegisterPlugins(
    NewEmotePlugin(),
    NewDicePlugin(),
    NewMusicPlugin(),
)
bot.Run()
```

---

## 🛠 贡献

1. Fork 本仓库
2. 新增插件到 `plugins/` 目录
3. 提交 PR 并说明功能

---

## 📄 License

MIT License © 月离
