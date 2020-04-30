FROM krancour/go-tools:v0.1.0
ARG LDFLAGS
ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/brigadecore/brigade
COPY brigade-cr-gateway/ brigade-cr-gateway/
COPY pkg/ pkg/
COPY vendor/ vendor/
RUN go build -ldflags "$LDFLAGS" -o bin/brigade-cr-gateway ./brigade-cr-gateway/cmd/brigade-cr-gateway
RUN mkdir /scratch-tmp

FROM scratch
# The glog library will write to here.
COPY --from=0 /scratch-tmp/ /tmp/
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /go/src/github.com/brigadecore/brigade/bin/brigade-cr-gateway /usr/bin/brigade-cr-gateway
CMD ["/usr/bin/brigade-cr-gateway"]
