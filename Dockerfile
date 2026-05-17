# syntax=docker/dockerfile:1

# go.mod Go 버전과 베이스를 맞춰 toolchain 자동 다운로드(GOTOOLCHAIN) 비용을 없앱니다.
FROM golang:1.25-alpine AS builder
WORKDIR /src

# private module / replace git URL 없음 — git 생략으로 apk·레이어 시간 절약
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
	go mod download

# 빌드에 필요한 소스만 복사해 레이어 캐시 hit 범위를 넓힙니다.
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY docs/ ./docs/

RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=0 GOOS=linux go build \
	-trimpath \
	-ldflags="-s -w" \
	-buildvcs=false \
	-o /server \
	./cmd/server

# 런타임: 작은 이미지 + 비루트 (TLS 연결용 CA 포함)
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata && \
	addgroup -g 65532 -S app && adduser -u 65532 -S app -G app

WORKDIR /
COPY --from=builder /server /server

USER app:app
EXPOSE 8080

ENTRYPOINT ["/server"]
