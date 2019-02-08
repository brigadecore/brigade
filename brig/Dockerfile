FROM alpine:3.8

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/amd64/brig /usr/bin/brig

CMD /usr/bin/brig
