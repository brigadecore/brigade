FROM ${DOCKER_ARCH}/alpine:3.8
${QEMU_IMPORT}

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/${ARCH}/brigade-controller /usr/bin/brigade-controller

CMD /usr/bin/brigade-controller
