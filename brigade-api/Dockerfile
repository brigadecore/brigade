FROM {{.docker}}/alpine:3.8
{{.qemu}}

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/{{.arch}}/brigade-api /usr/bin/brigade-api

CMD /usr/bin/brigade-api
