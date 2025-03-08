FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/oven/bun:latest AS builder

WORKDIR /build
COPY web/package.json .
RUN bun install
COPY ./web .
COPY ./VERSION .
RUN DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat VERSION) bun run build

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/golang:alpine3.21 AS builder2

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY=https://goproxy.cn

WORKDIR /build

ADD go.mod go.sum ./
RUN go mod tidy
RUN go mod download

COPY . .
COPY --from=builder /build/dist ./web/dist
RUN go mod tidy
RUN go build -ldflags "-s -w -X 'one-api/common.Version=$(cat VERSION)'" -o one-api

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/alpine:latest

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk update \
    && apk upgrade \
    && apk add --no-cache ca-certificates tzdata ffmpeg \
    && update-ca-certificates

COPY --from=builder2 /build/one-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]