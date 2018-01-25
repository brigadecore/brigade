FROM alpine:3.7

RUN apk update && apk add --no-cache \
    ca-certificates \
    git \
    && update-ca-certificates

COPY rootfs/brigade-github-gateway /usr/bin/brigade-github-gateway
COPY rootfs/gitssh.sh /gitssh.sh

ENV GIT_SSH=/gitssh.sh

CMD /usr/bin/brigade-github-gateway
