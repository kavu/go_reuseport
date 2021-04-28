test:
	@go test -v -cover

lint:
	@golangci-lint run

.PHONY: test lint
