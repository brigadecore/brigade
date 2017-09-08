FROM ubuntu:16.04

RUN apt-get update -y && apt-get install -y git

COPY rootfs /
ENV GIT_SSH=/gitssh.sh
ENV GIT_ASKPASS=/askpass.sh
CMD /clone.sh
