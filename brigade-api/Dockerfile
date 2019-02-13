FROM ${DOCKER_ARCH}/alpine:3.8
${QEMU_IMPORT}

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/${ARCH}/brigade-api /usr/bin/brigade-api

CMD /usr/bin/brigade-api
