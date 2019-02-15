FROM {{.docker}}/alpine:3.8
{{.qemu}}

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/{{.arch}}/brig /usr/bin/brig

CMD /usr/bin/brig
