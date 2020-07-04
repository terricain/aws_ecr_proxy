FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY aws_ecr_proxy /aws_ecr_proxy

ENV LISTEN_PORT=8080
ENV LISTEN_HOST=0.0.0.0
ENV LOG_LEVEL=INFO

CMD ["/aws_ecr_proxy"]
