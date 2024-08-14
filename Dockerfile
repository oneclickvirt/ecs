# syntax=docker/dockerfile:1

# 为支持的平台使用 Alpine
FROM --platform=$BUILDPLATFORM alpine:latest AS base-alpine

# 为 MIPS 平台使用替代基础镜像
FROM --platform=linux/mips debian:stable-slim AS base-mips
FROM --platform=linux/mipsle debian:stable-slim AS base-mipsle

# 选择适当的基础镜像
FROM base-$TARGETARCH$TARGETVARIANT AS final

# 安装必要的工具（需要根据基础镜像调整）
RUN if [ -f /etc/alpine-release ]; then \
        apk add --no-cache wget curl bash; \
    else \
        apt-get update && apt-get install -y wget curl bash; \
    fi

# 设置 GitHub URL 环境变量
ENV GITHUB_URL="https://github.com/oneclickvirt/ecs/releases/latest"

# 下载并执行 goecs.sh 脚本
RUN curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && \
    chmod +x goecs.sh && \
    bash goecs.sh env && \
    bash goecs.sh install

# 设置 goecs 为入口点
ENTRYPOINT ["goecs"]
