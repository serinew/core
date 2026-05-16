# syntax=docker/dockerfile:1

# go.mod 의 Go 버전보다 낮은 베이스를 쓸 경우, 빌드 시 toolchain 자동 준비(GOTOOLCHAIN=auto).
FROM golang:1.24-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
	-trimpath \
	-ldflags="-s -w" \
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
