FROM debian:jessie-slim
RUN apt-get update && apt-get install -y uuid-runtime curl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
  && mv kubectl /usr/local/bin/kubectl && chmod 755 /usr/local/bin/kubectl
COPY ./cron-event.sh /usr/local/bin/cron-event.sh
CMD /usr/local/bin/cron-event.sh
