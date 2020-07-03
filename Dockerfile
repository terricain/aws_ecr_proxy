FROM scratch
COPY aws_ecr_proxy /aws_ecr_proxy

CMD ["/aws_ecr_proxy"]
