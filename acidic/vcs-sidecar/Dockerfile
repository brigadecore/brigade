FROM ubuntu:16.04

RUN apt-get update -y && apt-get install -y git bzr mercurial subversion

COPY rootfs/gitssh.sh /gitssh.sh
COPY rootfs/vcs-sidecar /vcs-sidecar
ENV GIT_SSH=/gitssh.sh
CMD /vcs-sidecar
