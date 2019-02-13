FROM ${DOCKER_ARCH}/alpine:3.8
${QEMU_IMPORT}

RUN apk update && apk add --no-cache \
    ca-certificates \
    git \
    && update-ca-certificates

COPY rootfs/${ARCH}/brigade-github-gateway /usr/bin/brigade-github-gateway
COPY rootfs/gitssh.sh /gitssh.sh

ENV GIT_SSH=/gitssh.sh

CMD /usr/bin/brigade-github-gateway
