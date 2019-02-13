FROM ${DOCKER_ARCH}/alpine:3.8
${QEMU_IMPORT}

RUN apk update && apk add --no-cache \
    ca-certificates \
    git \
    && update-ca-certificates

COPY rootfs/${ARCH}/brigade-cr-gateway /usr/bin/brigade-cr-gateway

CMD /usr/bin/brigade-cr-gateway
