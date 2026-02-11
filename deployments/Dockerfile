# ========== 构建阶段 ==========
FROM golang:1.22-alpine AS builder

# 设置国内代理（加速依赖下载）
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0

WORKDIR /app

# 先复制依赖文件（利用 Docker 缓存层）
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并编译
COPY . .
RUN go build -ldflags="-s -w" -o /sentinel ./cmd/sentinel/

# ========== 运行阶段 ==========
FROM alpine:3.19

# 安装 CA 证书（HTTPS 请求需要）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区（根据需要调整）
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /sentinel .

# 非 root 用户运行（安全最佳实践）
RUN adduser -D -u 1000 sentinel
USER sentinel

ENTRYPOINT ["./sentinel"]
