FROM quay.io/deis/lightweight-docker-go:v0.7.0
ARG LDFLAGS
ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/brigadecore/brigade
COPY brigade-api/ brigade-api/
COPY pkg/ pkg/
COPY vendor/ vendor/
RUN go build -ldflags "$LDFLAGS" -o bin/brigade-api ./brigade-api/cmd/brigade-api
RUN mkdir /scratch-tmp

FROM scratch
# The glog library will write to here.
COPY --from=0 /scratch-tmp/ /tmp/
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /go/src/github.com/brigadecore/brigade/bin/brigade-api /usr/bin/brigade-api
CMD ["/usr/bin/brigade-api"]
