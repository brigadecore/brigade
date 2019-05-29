FROM quay.io/deis/lightweight-docker-go:v0.6.0
ARG LDFLAGS
ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/brigadecore/brigade
COPY brigade-cr-gateway/ brigade-cr-gateway/
COPY pkg/ pkg/
COPY vendor/ vendor/
RUN go build -ldflags "$LDFLAGS" -o bin/brigade-cr-gateway ./brigade-cr-gateway/cmd/brigade-cr-gateway

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /go/src/github.com/brigadecore/brigade/bin/brigade-cr-gateway /usr/bin/brigade-cr-gateway
# We have dependencies that use glog, which tries to write logs to /tmp/
# Scratch containers don't HAVE a /tmp/ directory-- at least not by default.
# We redirect the log output to stderr using the --logtostderr switch.
CMD ["/usr/bin/brigade-cr-gateway", "--logtostderr"]
