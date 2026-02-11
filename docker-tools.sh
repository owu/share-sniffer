#!/bin/bash

# 定义常量
CONFIG_FILE="./internal/config/config.go"

# 读取版本号
if [ -f "$CONFIG_FILE" ]; then
    VERSION=$(grep -E 'q\.AppInfo\.Version\s*=\s*"[^"]+"' "$CONFIG_FILE" | sed -E 's/.*"([^"]+)".*/\1/' | tr -d '[:space:]')
    if [ -z "$VERSION" ]; then
        echo "警告：无法从 $CONFIG_FILE 中读取版本号"
        exit 1
    fi
else
    echo "警告：未找到配置文件 $CONFIG_FILE，版本号读取失败"
    exit 1
fi

echo "版本号: $VERSION"

IMAGE_NAME="share-sniffer"
FULL_IMAGE_NAME="${IMAGE_NAME}:${VERSION}"
TAR_FILE="${IMAGE_NAME}.${VERSION}.tar"

# 显示帮助信息
show_help() {
    echo "使用方法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  b           构建 Docker 镜像 (可选输入代理地址)"
    echo "  d           停止 Docker 容器"
    echo "  l           查看容器日志"
    echo "  m           镜像迁移 (导出/导入)"
    echo "  u           启动 Docker 容器"
    echo "  h           显示此帮助信息"
}

# 构建 Docker 镜像
build_image() {
    echo "========================================"
    echo "开始构建 Docker 镜像"
    echo "镜像名称：${FULL_IMAGE_NAME}"
    echo "请输入代理地址 (例如 http://192.168.1.10:10808) [可选，直接回车不使用代理]:"
    echo "========================================"
    
    read -p "代理地址: " proxy_url
    
    if [ -z "$proxy_url" ]; then
        echo "未输入代理地址，将进行无代理构建..."
        if docker build -t "${FULL_IMAGE_NAME}" .; then
            echo "✓ Docker 镜像构建成功"
        else
            echo "✗ Docker 镜像构建失败"
            return 1
        fi
    else
        echo "使用代理地址: $proxy_url 进行构建..."
        if docker build \
            --network host \
            --build-arg HTTP_PROXY="$proxy_url" \
            --build-arg HTTPS_PROXY="$proxy_url" \
            -t "${FULL_IMAGE_NAME}" .; then
            echo "✓ Docker 镜像构建成功 (带代理)"
        else
            echo "✗ Docker 镜像构建失败"
            return 1
        fi
    fi
}

# 启动 Docker 容器
start_containers() {
    echo "启动 Docker 容器服务..."
    
    # 确保目录存在
    mkdir -p ./logs
    mkdir -p ./config

    # 使用 docker-compose 启动
    if [ -f "docker-compose.yml" ]; then
        # 更新 compose 文件中的镜像版本（如果需要，这里可以添加 sed 命令来替换镜像 tag，但通常 compose 文件手动维护）
        # 这里仅打印提示
        echo "注意：请确保 docker-compose.yml 中的镜像版本与当前构建版本 (${VERSION}) 一致"
        docker compose up -d
    else
        # 如果没有 compose 文件，使用 docker run
        echo "未找到 docker-compose.yml，尝试使用 docker run 启动..."
        docker run -d \
            --name "${IMAGE_NAME}" \
            --network host \
            -v "$(pwd)/logs:/app/logs" \
            --restart always \
            "${FULL_IMAGE_NAME}"
    fi
    echo "启动命令已执行，请使用 'docker ps' 查看状态"
}

# 停止 Docker 容器
stop_containers() {
    echo "停止 Docker 容器服务..."
    if [ -f "docker-compose.yml" ]; then
        docker compose down
    else
        docker stop "${IMAGE_NAME}" && docker rm "${IMAGE_NAME}"
    fi
}

# 查看日志
show_logs() {
    docker logs -f "${IMAGE_NAME}"
}

# 镜像迁移
migration_menu() {
    echo "1) 导出镜像 (docker save)"
    echo "2) 导入镜像 (docker load)"
    read -p "请选择 [1/2]: " -r choice
    case "$choice" in
        1) docker save -o "${TAR_FILE}" "${FULL_IMAGE_NAME}" && echo "导出成功: ${TAR_FILE}" ;;
        2) docker load -i "${TAR_FILE}" && echo "导入成功" ;;
        *) echo "无效选择" ;;
    esac
}

# 主逻辑
case "$1" in
    b) build_image ;;
    d) stop_containers ;;
    l) show_logs ;;
    m) migration_menu ;;
    u) start_containers ;;
    h|*) show_help ;;
    esac
