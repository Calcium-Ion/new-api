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
RUN go mod download

COPY . .
COPY --from=builder /build/dist ./web/dist
RUN go build -ldflags "-s -w -X 'one-api/common.Version=$(cat VERSION)'" -o one-api

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/library/a