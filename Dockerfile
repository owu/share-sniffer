# 第一阶段：构建阶段
FROM golang:1.24 AS builder

# 声明构建参数
ARG HTTP_PROXY
ARG HTTPS_PROXY

# 将参数设置为环境变量，供容器内的程序使用
ENV HTTP_PROXY=${HTTP_PROXY}
ENV HTTPS_PROXY=${HTTPS_PROXY}

# 设置工作目录
WORKDIR /app

# 允许 Go 自动下载需要的工具链版本
ENV GOTOOLCHAIN=auto

# 设置 Go 代理以加快下载速度
ENV GOPROXY=https://goproxy.cn,direct

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 屏蔽依赖chromedp的类型
ENV IS_DESKTOP=0

# 编译项目
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o share-sniffer ./launcher/api/main.go
# 编译CLI工具
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o share-sniffer-cli ./launcher/cli/main.go

# 第二阶段：运行阶段
FROM alpine:latest

# 替换为阿里云镜像源，解决代理连接超时问题
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的运行时库和工具
# ca-certificates: HTTPS证书
# tzdata: 时区数据
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai
ENV IS_DESKTOP=0
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 声明构建参数
ARG HTTP_PROXY
ARG HTTPS_PROXY

# 将参数设置为环境变量，供容器内的程序使用
ENV HTTP_PROXY=${HTTP_PROXY}
ENV HTTPS_PROXY=${HTTPS_PROXY}

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/share-sniffer /app/share-sniffer
COPY --from=builder /app/share-sniffer-cli /app/bin/share-sniffer-cli

# 复制默认配置文件到镜像中
COPY config/dockerconfig.toml /app/config/config.toml

# 设置执行权限
RUN chmod +x /app/share-sniffer
RUN chmod +x /app/bin/share-sniffer-cli

# 声明端口 (仅作为文档说明，实际端口由配置文件./config/dockerconfig.toml决定)
EXPOSE 60204

# 启动程序
CMD ["/app/share-sniffer", "-config", "/app/config/config.toml"]
