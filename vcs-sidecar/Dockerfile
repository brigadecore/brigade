FROM ubuntu:16.04

RUN apt-get update -y && apt-get install -y git bzr mercurial subversion

COPY rootfs /
ENV GIT_SSH=/gitssh.sh
CMD /vcs-sidecar
