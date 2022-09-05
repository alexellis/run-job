FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.16 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN mkdir -p /home/app

# Add non root user
RUN addgroup -S app && adduser app -S -G app
RUN chown app /home/app

RUN apk add --no-cache curl && curl -SLs https://get.arkade.dev | sh

WORKDIR /home/app
USER app

RUN arkade get kubectl@v1.24.1 --quiet

CMD ["/home/app/.arkade/bin/kubectl", "get", "nodes", "-o", "wide"]
