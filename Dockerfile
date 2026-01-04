# [Dockerfile]
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder
WORKDIR /app

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH \
    $( [ "$TARGETARCH" = "arm" ] && echo "GOARM=${TARGETVARIANT#v}" ) \
    go build -o g ./main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/g /usr/local/bin/g
ENTRYPOINT ["g"]
