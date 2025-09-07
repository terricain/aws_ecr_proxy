FROM --platform=$BUILDPLATFORM golang:1.24.7-alpine3.22 AS build

WORKDIR /usr/local/go/src/aws_ecr_proxy

COPY go.mod go.sum /usr/local/go/src/aws_ecr_proxy/

RUN go mod download

COPY cmd/ /usr/local/go/src/aws_ecr_proxy/cmd/
COPY internal/ /usr/local/go/src/aws_ecr_proxy/internal/

ENV PKG=github.com/terricain/aws_ecr_proxy
ARG DOCKER_METADATA_OUTPUT_JSON
ARG TARGETOS
ARG TARGETARCH

# hadolint ignore=DL3018,SC2086,DL4006,SC2155
RUN apk add --no-cache jq && \
    export VERSION="$(echo "${DOCKER_METADATA_OUTPUT_JSON}" | jq -r '.labels["org.opencontainers.image.version"]')" && \
    export GIT_COMMIT="$(echo "${DOCKER_METADATA_OUTPUT_JSON}" | jq -r '.labels["org.opencontainers.image.revision"]')" && \
    export BUILD_DATE="$(echo "${DOCKER_METADATA_OUTPUT_JSON}" | jq -r '.labels["org.opencontainers.image.created"]')" && \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
      -o /server \
      -ldflags "-X ${PKG}/internal/version.VERSION=${VERSION} -X ${PKG}/internal/version.SHA=${GIT_COMMIT} -X ${PKG}/internal/version.BUILDDATE=${BUILD_DATE} -s -w" \
      cmd/aws_ecr_proxy/main.go

FROM gcr.io/distroless/static-debian12:nonroot AS release
COPY --from=build /server /aws_ecr_proxy

ENV LISTEN_PORT=8080
ENV LISTEN_HOST=0.0.0.0
ENV LOG_LEVEL=INFO

CMD ["/aws_ecr_proxy"]
