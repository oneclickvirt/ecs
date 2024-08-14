# syntax=docker/dockerfile:1

FROM alpine:latest

# 安装必要的工具
RUN apk add --no-cache wget curl bash

# 下载并执行 goecs.sh 脚本
RUN curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && \
    chmod +x goecs.sh && \
    bash goecs.sh env && \
    bash goecs.sh install

# 设置 goecs 为入口点
ENTRYPOINT ["goecs"]