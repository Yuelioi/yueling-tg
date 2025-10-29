# -------------------- 构建阶段 --------------------
FROM golang:1.25-alpine AS builder

# 安装 CGO 所需依赖
RUN apk add --no-cache \
  gcc g++ musl-dev pkgconfig libwebp-dev git

WORKDIR /app

# 先复制模块文件并下载依赖（缓存友好）
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译二进制（启用 CGO）
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o main .

# -------------------- 运行阶段 --------------------
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache libwebp ca-certificates tzdata

ENV TZ=Asia/Shanghai
WORKDIR /app

# 复制构建好的二进制
COPY --from=builder /app/main .

# 使用非 root 用户运行
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
USER appuser

CMD ["./main"]
