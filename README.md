# AWS ECR Proxy

Simple ECR proxy which manages the ECR authentication and handles the Link headers.

The reason I created this was, FluxCD was not playing ball with ECR when ran outside of AWS, and the standard Nginx ECR proxies don't handle `Link` headers which Docker 
registries use for pagination, which results in Flux complaining about the registry requiring authentication.


....todo  