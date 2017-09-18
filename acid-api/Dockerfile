FROM alpine:3.6

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY rootfs/acid-api /usr/bin/acid-api

CMD /usr/bin/acid-api
