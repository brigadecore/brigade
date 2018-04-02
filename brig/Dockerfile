FROM alpine:3.7

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/brig /usr/bin/brig

CMD /usr/bin/brig