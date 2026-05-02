# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.25.8 AS builder

ARG TARGETOS
ARG TARGETARCH

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GIN_MODE=release \
  GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download
COPY . .
ARG VERSION=unknown
RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -trimpath -ldflags="-s -w -X main.Version=${VERSION}" -o wechat-robot-client


FROM registry.cn-shenzhen.aliyuncs.com/houhou/silk-base:latest

ENV GIN_MODE=release \
  TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/wechat-robot-client ./

EXPOSE 9000

ENTRYPOINT []
CMD ["/app/wechat-robot-client"]