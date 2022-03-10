FROM $TARGETARCH/alpine:3.15.0

ENV ARTEMIS_VERSION=2.19.0
ENV ARTEMIS_USERNAME=artemis
ENV ARTEMIS_PASSWORD=artemis

WORKDIR /opt/artemis
## Special note about the URL for downloading Artemis: Annoyingly, the URL for
## the latest release and the the URLs for previous releases do not vary only
## by version number. If this build starts breaking, the very likely cause is a
## 404 on the download. It can be fixed by upgrading to the latest Artemis or
## by changing the download URL.
RUN buildDeps="curl" && \ 
  apk update && \
  apk add \
    $buildDeps \
    openjdk11-jre \
    xmlstarlet && \
  curl https://dlcdn.apache.org/activemq/activemq-artemis/${ARTEMIS_VERSION}/apache-artemis-${ARTEMIS_VERSION}-bin.tar.gz \
    -o apache-artemis-${ARTEMIS_VERSION}-bin.tar.gz && \
  tar xzf apache-artemis-${ARTEMIS_VERSION}-bin.tar.gz --strip 1 && \
  apk del $buildDeps && \
  mkdir /var/lib/artemis && \
  addgroup -S -g 1000 artemis && \
  adduser -S -u 1000 artemis -G artemis && \
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
