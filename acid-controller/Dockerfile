FROM alpine:3.6

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/acid-controller /usr/bin/acid-controller

CMD /usr/bin/acid-controller
