#!/bin/bash
set -e # エラーがあったら即終了

# 出力ディレクトリの作成
mkdir -p bin

echo "Building for Linux (amd64)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/linux-amd64/g ./main.go

echo "Building for Linux (arm64)..."
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/linux-arm64/g ./main.go

echo "Building for Linux (arm/v7)..."
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/linux-arm/g ./main.go

echo "Building for macOS..."
GOOS=darwin CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/darwin/g ./main.go

echo "Building for Windows..."
GOOS=windows CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/windows/g ./main.go

echo "Build complete."
