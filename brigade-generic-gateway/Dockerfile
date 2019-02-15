FROM {{.docker}}/alpine:3.8
{{.qemu}}

RUN apk update && apk add --no-cache \
  ca-certificates \
  git \
  && update-ca-certificates

COPY rootfs/{{.arch}}/brigade-generic-gateway /usr/bin/brigade-generic-gateway

CMD /usr/bin/brigade-generic-gateway
