FROM alpine:3.8

RUN apk update && apk add --no-cache \
    ca-certificates \
    git \
    && update-ca-certificates

COPY rootfs/amd64/brigade-generic-gateway /usr/bin/brigade-generic-gateway

CMD /usr/bin/brigade-generic-gateway
