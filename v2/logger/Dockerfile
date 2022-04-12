# This Dockerfile builds a custom Fluentd image that layers in the Fluentd
# plugins we rely upon on top of a standard Fluentd image.
FROM fluentd:v1.14.0-1.0

USER root

RUN buildDeps="make gcc g++ libc-dev libffi-dev ruby-dev" \
  && apk update \
  && apk add \
    $buildDeps \
    libffi \
    net-tools \
  && gem install \
    bson:4.12.1 \
    fluent-plugin-rewrite-tag-filter:2.4.0 \
    fluent-plugin-mongo:1.5.0 \
    fluent-plugin-kubernetes_metadata_filter:2.9.1 \
    fluent-plugin-multi-format-parser:1.0.0 \
  && apk del $buildDeps \
  && gem sources --clear-all \
  && rm -rf /tmp/* /var/tmp/* /usr/lib/ruby/gems/*/cache/*.gem
