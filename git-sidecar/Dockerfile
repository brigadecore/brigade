FROM alpine:3.7

RUN apk update && apk add --no-cache \
    ca-certificates \
    git \
    openssh-client \
    && update-ca-certificates

COPY rootfs /
ENV GIT_SSH=/gitssh.sh
ENV GIT_ASKPASS=/askpass.sh
CMD /clone.sh
