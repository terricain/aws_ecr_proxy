lint:
	go fmt ./...
	golangci-lint run
	go mod tidy
