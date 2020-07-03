now = $(shell date +'%Y-%m-%dT%T')
version = $(word 3, $(subst /, ,${GITHUB_REF}))

build:
	CGO_ENABLED=0 go build -o aws_ecr_proxy -ldflags="-s -w -X github.com/terrycain/aws_ecr_proxy/internal/version.SHA=${GITHUB_SHA} -X github.com/terrycain/aws_ecr_proxy/internal/version.BUILDDATE=${now} -X github.com/terrycain/aws_ecr_proxy/internal/version.VERSION=${version}" cmd/aws_ecr_proxy/main.go

lint:
	go fmt ./...
	golangci-lint run
	go mod tidy
