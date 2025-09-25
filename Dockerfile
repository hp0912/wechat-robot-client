FROM golang:1.23 AS builder

# 定义版本号参数
ARG VERSION=unknown

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GIN_MODE=release \
  GOPROXY=https://goproxy.cn,direct

WORKDIR /app
ADD go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w -X main.Version=${VERSION}" -o wechat-robot-client


FROM registry.cn-shenzhen.aliyuncs.com/houhou/silk-base:latest

ENV GIN_MODE=release \
  TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/wechat-robot-client ./

EXPOSE 9000

ENTRYPOINT []
CMD ["/app/wechat-robot-client"]