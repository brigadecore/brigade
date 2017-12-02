FROM alpine:3.7

RUN apk update && apk add --no-cache \
    ca-certificates \
    git \
    && update-ca-certificates

COPY rootfs/brigade-cr-gateway /usr/bin/brigade-cr-gateway

CMD /usr/bin/brigade-cr-gateway
