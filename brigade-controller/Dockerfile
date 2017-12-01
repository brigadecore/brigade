FROM alpine:3.7

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/brigade-controller /usr/bin/brigade-controller

CMD /usr/bin/brigade-controller
