# -------------------- 构建阶段 --------------------
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 复制依赖文件并预下载模块（缓存友好）
COPY go.mod go.sum ./
RUN go mod download

# 复制项目源码
COPY . .

# 使用纯 Go 静态编译（无CGO）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# -------------------- 运行阶段 --------------------
FROM alpine:latest

# 安装运行时基础依赖
RUN apk add --no-cache \
  ca-certificates \
  tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 复制编译好的二进制
COPY --from=builder /app/main .

# 使用安全的非 root 用户运行
USER nobody

CMD ["./main"]
