# -------------------- 构建阶段 --------------------
FROM golang:1.25-alpine AS builder

# 安装构建依赖（CGO 所需）
RUN apk add --no-cache gcc g++ musl-dev pkgconfig libwebp-dev git

# 设置工作目录
WORKDIR /app

# 先复制 go.mod go.sum 并下载模块，利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译二进制（启用 CGO）
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o main .

# -------------------- 运行阶段 --------------------
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache ca-certificates libwebp tzdata

ENV TZ=Asia/Shanghai

WORKDIR /app

# 复制构建好的二进制
COPY --from=builder /app/main .

# 使用安全的非 root 用户
USER nobody

CMD ["./main"]
