FROM {{.docker}}/alpine:3.8
{{.qemu}}

COPY rootfs/{{.arch}}/brigade-vacuum /usr/bin/brigade-vacuum

CMD /usr/bin/brigade-vacuum
