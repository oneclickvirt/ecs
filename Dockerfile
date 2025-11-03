# syntax=docker/dockerfile:1
FROM alpine:latest

RUN apk update && apk add --no-cache wget curl bash || \
    (echo "Standard repo failed, trying edge repo..." && \
     apk add --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main wget curl bash)

RUN apk add --no-cache bind-tools || \
    (echo "Standard repo failed for bind-tools, trying edge repo..." && \
     apk add --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main bind-tools)

RUN apk add --no-cache grep openssl ca-certificates || \
    (echo "Standard repo failed, trying edge repo..." && \
     apk add --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main grep openssl ca-certificates)

RUN apk add --no-cache uuidgen || \
    apk add --no-cache util-linux || \
    (echo "Standard repo failed for uuidgen, trying edge repo..." && \
     apk add --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main uuidgen) || \
    apk add --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main util-linux

RUN export noninteractive=true
# 下载并执行 goecs.sh 脚本
RUN curl -L https://raw.githubusercontent.com/oneclickvirt/ecs/master/goecs.sh -o goecs.sh && \
    chmod +x goecs.sh && \
    bash goecs.sh env && \
    bash goecs.sh install
# 设置 goecs 为入口点
ENTRYPOINT ["goecs"]
