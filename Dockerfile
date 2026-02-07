FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
WORKDIR /app

ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0
ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w" -o gaproxy ./main.go

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /app/gaproxy /usr/local/bin/gaproxy

ENTRYPOINT ["gaproxy"]
