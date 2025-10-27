# -------------------- 构建阶段 --------------------
FROM golang:1.25-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache \
  git \
  gcc \
  g++ \
  musl-dev \
  pkgconfig \
  ffmpeg-dev \
  freetype-dev

# 设置工作目录
WORKDIR /app

# 复制依赖文件，加快缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建可执行文件（启用 CGO 支持）
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# -------------------- 运行阶段 --------------------
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
  ca-certificates \
  freetype \
  tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
  adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 复制构建阶段的二进制文件
COPY --from=builder /app/main .

# 切换到非 root 用户
USER appuser

# 启动命令
CMD ["./main"]
