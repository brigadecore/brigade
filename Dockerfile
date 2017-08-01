FROM debian:jessie-slim

RUN apt-get update && apt-get install -y git

COPY rootfs /

ENV GIT_SSH=/gitssh.sh

CMD /usr/bin/acid
