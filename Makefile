fmt:
	go mod tidy
	go fix -diff ./...
	golangci-lint run --fix ./examples/... ./naive/...

test-naive:
	go clean -testcache
	go test ./naive -race -timeout=5m -v -count=10

test-gnaive:
	go clean -testcache
	go test ./gnaive -race -timeout=5m -v -count=10

cover-naive:
	go test ./naive -race -cover

cover-gnaive:
	go test ./gnaive -race -cover


run-naive:
	go run ./examples/naive

run-gnaive:
	go run ./examples/gnaive

.PHONY: fmt test-naive test-gnaive cover-naive cover-gnaive run-naive run-gnaive
