FROM alpine:3.8

RUN apk update && apk add --no-cache \
    ca-certificates \
    git \
    openssh-client \
    && update-ca-certificates

COPY git-sidecar/rootfs/ /
ENV GIT_SSH=/gitssh.sh
ENV GIT_ASKPASS=/askpass.sh
CMD /clone.sh
