FROM debian:jessie-slim

RUN apt-get update && apt-get install -y git curl \
		&& curl -L https://storage.googleapis.com/kubernetes-release/release/v1.6.1/bin/linux/amd64/kubectl -o /usr/bin/kubectl \
		&& chmod +x /usr/bin/kubectl

COPY acid /usr/bin/acid
COPY gitssh.sh /gitssh.sh

ENV GIT_SSH=/gitssh.sh

CMD /usr/bin/acid
