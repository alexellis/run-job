FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.16 as builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN mkdir -p /home/app

RUN apk add --no-cache nodejs npm

# Add non root user
RUN addgroup -S app && adduser app -S -G app
RUN chown app /home/app

WORKDIR /home/app
USER app

COPY package.json    .
RUN npm install --omit=dev
COPY index.js        .

CMD ["/usr/bin/node", "./index.js"]
