FROM ${DOCKER_ARCH}/alpine:3.8
${QEMU_IMPORT}

COPY rootfs/${ARCH}/brigade-vacuum /usr/bin/brigade-vacuum

CMD /usr/bin/brigade-vacuum
