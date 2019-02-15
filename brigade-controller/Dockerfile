FROM {{.docker}}/alpine:3.8
{{.qemu}}

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/{{.arch}}/brigade-controller /usr/bin/brigade-controller

CMD /usr/bin/brigade-controller
