FROM openjdk:8-jre-bullseye

ARG ARTEMIS_VERSION=2.20.0

ENV ARTEMIS_USERNAME=artemis
ENV ARTEMIS_PASSWORD=artemis

WORKDIR /opt/artemis
## Special note about the URL for downloading Artemis: Annoyingly, the URL for
## the latest release and the the URLs for previous releases do not vary only
## by version number. If this build starts breaking, the very likely cause is a
## 404 on the download. It can be fixed by upgrading to the latest Artemis or
## by changing the download URL.
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    libaio1 \
    xmlstarlet &&\ 
  rm -rf /var/lib/apt/lists/* && \
  curl https://dlcdn.apache.org/activemq/activemq-artemis/${ARTEMIS_VERSION}/apache-artemis-${ARTEMIS_VERSION}-bin.tar.gz \
    -o apache-artemis-${ARTEMIS_VERSION}-bin.tar.gz && \
  tar xzf apache-artemis-${ARTEMIS_VERSION}-bin.tar.gz --strip 1 && \
  rm apache-artemis-${ARTEMIS_VERSION}-bin.tar.gz && \
  mkdir /var/lib/artemis && \
  groupadd -g 1000 -r artemis && \
  useradd -r -u 1000 -g artemis artemis && \
  chown -R artemis.artemis /var/lib/artemis

COPY v2/artemis/docker-entrypoint.sh /

COPY v2/artemis/merge.xslt /var/lib/artemis/assets/

# Web UI
EXPOSE 8161
# CORE
EXPOSE 61616
# AMQP
EXPOSE 5672

USER artemis
WORKDIR /var/lib/artemis

ENTRYPOINT [ "/docker-entrypoint.sh" ]
CMD ["artemis-server"]
