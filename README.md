# AWS ECR Proxy

Simple ECR proxy which manages AWS ECR authentication and handles the Link headers. 
The container also has endpoints for Kubernetes liveness and readyness probes.

### Usage

#### CLI Example

Example usage:
```bash
docker run -e AWS_REGION=eu-west-1 \
           -e AWS_SECRET_ACCESS_KEY=blah \
           -e AWS_ACCESS_KEY_ID=blah \
           --name registry --rm -i \
            -p 8080:8080 terrycain/aws_ecr_proxy:latest
```

#### Environment Variables

* `AWS_REGION` - Confiures the AWS SDK's region. This will determine which regions ECR images are available
* `AWS_ACCESS_KEY_ID` - AWS Access Key
* `AWS_SECRET_ACCESS_KEY` - AWS Secret Key
* `LOG_LEVEL` - Default `INFO` - Sets the logging level, one of: `DEBUG`, `INFO`, `WARN`, `ERROR`
* `LISTEN_PORT` - Default `8080`
* `LISTEN_HOST` - Default `0.0.0.0`

This proxy uses the standard AWS SDK, so it is entirely possible the AWS specific environment variables 
can be omitted and the proxy should attempt to authenticate using an appropriate IAM role, but this is untested.


### How it works

On startup, the proxy will start off a loop to grab an ECR token and continuously renew it roughly every 12 hours (unless amazon change the expiry).

On request, it'll inject an Authorization header containing the ECR token. Before serving ECR's response it will
modify any `Link` headers which are used for pagination and contain ECR urls; the header will have its links updated with links referencing the proxy.

### Why

The reason I created this was, FluxCD was not playing ball with ECR when ran outside of AWS, and the standard NGINX ECR proxies don't handle `Link` headers which Docker 
registries use for pagination, which results in Flux complaining about the registry requiring authentication. Until the pagination kicked in the standard proxy `https://github.com/catalinpan/aws-ecr-proxy` worked fine.

### Todo

* add support to listen with TLS
* Tests :/
