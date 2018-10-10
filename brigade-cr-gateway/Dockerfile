FROM debian:jessie-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    git \
    && rm -rf /var/lib/apt/lists/*

COPY rootfs/brigade-cr-gateway /usr/bin/brigade-cr-gateway

CMD /usr/bin/brigade-cr-gateway
