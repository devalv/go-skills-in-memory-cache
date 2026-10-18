fmt:
	go mod tidy
	go fix -diff ./...
	golangci-lint run --fix ./examples/... ./naive/...

test-basic:
	go test ./... -race -timeout=5m -v

cover-basic:
	go test ./... -race -cover

run-basic:
	go run ./examples/basic

.PHONY: fmt test cover run-basic
