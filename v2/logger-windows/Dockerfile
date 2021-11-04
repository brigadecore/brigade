# This Dockerfile builds a custom Fluentd image that layers in the Fluentd
# plugins we rely upon on top of a standard Fluentd image.

FROM fluent/fluentd:v1.14.2-windows-ltsc2019-1.0

RUN gem install bson -v 4.12.0 \
  && gem install fluent-plugin-rewrite-tag-filter -v 2.4.0 \
  && gem install fluent-plugin-mongo -v 1.5.0 \
  && gem install fluent-plugin-kubernetes_metadata_filter -v 2.9.1 \
  && gem install fluent-plugin-multi-format-parser -v 1.0.0 \
  && gem sources --clear-all \
  && powershell -Command "rm -fo C:\ruby26\lib\ruby\gems\2.6.0\cache\*.gem"
