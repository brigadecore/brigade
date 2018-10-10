FROM debian:jessie-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY rootfs/brigade-api /usr/bin/brigade-api

CMD /usr/bin/brigade-api
