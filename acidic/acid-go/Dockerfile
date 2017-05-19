FROM golang:1.7

COPY rootfs/hook.sh /hook.sh
COPY rootfs/main.sh /hook/data/main.sh
COPY rootfs/gitssh.sh /gitssh.sh
ENV GIT_SSH=/gitssh.sh
CMD /hook.sh
