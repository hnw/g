GOOS=darwin go build -o ./bin/mac/gaproxy -ldflags="-s -w" ./main.go
GOARCH=amd64 GOOS=linux go build -o ./bin/linux/gaproxy -ldflags="-s -w" ./main.go
GOOS=windows go build -o ./bin/windows/gaproxy -ldflags="-s -w" ./main.go
