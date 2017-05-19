FROM ubuntu:16.04

RUN apt-get update -y && apt-get install -y git make

COPY rootfs/hook.sh /hook.sh
COPY rootfs/main.sh /hook/data/main.sh
COPY rootfs/gitssh.sh /gitssh.sh
ENV GIT_SSH=/gitssh.sh
CMD /hook.sh
