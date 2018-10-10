FROM debian:jessie-slim

RUN apt-get update && apt-get install -y \
  git \
  openssh-client \
  ca-certificates \
  && rm -rf /var/lib/apt/lists/*

COPY rootfs /
ENV GIT_SSH=/gitssh.sh
ENV GIT_ASKPASS=/askpass.sh
CMD /clone.sh
