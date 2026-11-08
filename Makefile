fmt:
	go mod tidy
	go fix -diff ./...
	golangci-lint run --fix ./...

test-naive:
	go clean -testcache
	go test ./naive -race -timeout=5m -v -count=10

test-gnaive:
	go clean -testcache
	go test ./gnaive -race -timeout=5m -v -count=10

test-ttl:
	go clean -testcache
	go test ./ttl -race -timeout=5m -v -count=10 -short
# 	go test ./ttl -race -timeout=5m -v -count=10 -short

cover-naive:
	go test ./naive -cover

cover-gnaive:
	go test ./gnaive -cover

cover-ttl:
	go test ./ttl -cover
# 	go test -coverprofile=coverage.out ./ttl
# 	go tool cover -func=coverage.out
# 	go tool cover -html=coverage.out

run-naive:
	go run ./examples/naive

run-gnaive:
	go run ./examples/gnaive

run-ttl:
	go run ./examples/ttl

test-lru:
	go clean -testcache
	go test ./lru -race -timeout=5m -v -count=10

cover-lru:
	go test ./lru -cover

run-lru:
	go run ./examples/lru

.PHONY: fmt test-naive test-gnaive cover-naive cover-gnaive run-naive run-gnaive test-ttl cover-ttl run-ttl test-lru cover-lru run-lru
