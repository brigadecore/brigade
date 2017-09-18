FROM alpine:3.6

RUN apk update && apk add --no-cache \
    ca-certificates \
    git \
    && update-ca-certificates

COPY rootfs /

ENV GIT_SSH=/gitssh.sh

CMD /usr/bin/acid-gateway
