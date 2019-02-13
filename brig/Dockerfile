FROM ${DOCKER_ARCH}/alpine:3.8
${QEMU_IMPORT}

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/${ARCH}/brig /usr/bin/brig

CMD /usr/bin/brig
