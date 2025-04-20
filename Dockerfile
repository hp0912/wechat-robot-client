FROM golang:1.23 AS builder

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GIN_MODE=release \
  GOPROXY=https://goproxy.cn,direct

WORKDIR /app
ADD go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o wechat-robot-client


FROM alpine:latest

ENV GIN_MODE=release \
  TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/wechat-robot-client ./

EXPOSE 9000

CMD ["/app/wechat-robot-client"]