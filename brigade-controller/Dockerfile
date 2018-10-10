FROM debian:jessie-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY rootfs/brigade-controller /usr/bin/brigade-controller

CMD /usr/bin/brigade-controller
