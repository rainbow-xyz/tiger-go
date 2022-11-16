FROM golang:1.18-alpine AS build

# 复制源码并执行build，此处当文件有变化会产生新的一层镜像层
COPY . /go/src/saas_service
WORKDIR /go/src/saas_service

RUN set -eux; \
      go mod download; \
      go env -w CGO_ENABLED=0; \
      go build -o /opt/saas/passport/bin/saas_passport_server /go/src/saas_service/cmd/passport/server.go;


FROM alpine:3.16.2 AS runtime

COPY build/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
RUN set -eux; \
      apk add curl; \
      # 设置时区
      ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime; \
      echo "Asia/Shanghai" > /etc/timezone;

FROM runtime AS saas-passport-server

RUN set -eux; \
      \
      # 创建日志目录
      mkdir -p /var/logs/saas/passport; \
      \
      # 创建配置文件目录
      mkdir -p /opt/saas/passport/configs;

# 测试版可以把配置打在镜像里，正式版单独配置，不要将配置打进镜像
# COPY internal/passport/configs /opt/saas/passport/configs
COPY --from=build /opt/saas/passport/bin/saas_passport_server /opt/saas/passport/bin/saas_passport_server
ENTRYPOINT ["/opt/saas/passport/bin/saas_passport_server"]