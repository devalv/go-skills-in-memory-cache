setup:
	go install golang.org/x/tools/cmd/goimports@latest
	go install mvdan.cc/gofumpt@latest

fmt:
	go mod tidy
	go fix -diff ./...
# 	golangci-lint run --fix ./cmd/... ./internal/...

test:
	go test ./... -race -timeout=5m -v

cover:
	go test ./... -race -cover

build:
	$(MAKE) fmt
	go env -w CGO_ENABLED=0
	go env -w GOOS=linux
	go env -w GOARCH=amd64
	go build -o application ./cmd/app

run:
	go run app.go

.PHONY: setup fmt test build cover run build-deb
