FROM quay.io/deis/lightweight-docker-go:v0.6.0
ARG LDFLAGS
ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/brigadecore/brigade
COPY brigade-vacuum/ brigade-vacuum/
COPY pkg/ pkg/
COPY vendor/ vendor/
RUN go build -ldflags "$LDFLAGS" -o bin/brigade-vacuum ./brigade-vacuum/cmd/brigade-vacuum
RUN mkdir /scratch-tmp

FROM scratch
# We have dependencies that use glog, which tries to write logs to /tmp/
# Scratch containers don't HAVE a /tmp/ directory-- at least not by default.
# Ordinarily, we can redirect those logs using a --logtostderr switch, but this
# binary has its OWN switch parsing that will not recognize that switch and
# will therefore fail. We also can't mkdir here because scratch containers don't
# have the mkdir binary. We fall back on something super hacky-- we create a
# new, empty directory in the first stage of the build and copy it over to
# /tmp/ in our new container.
COPY --from=0 /scratch-tmp/ /tmp/
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=0 /go/src/github.com/brigadecore/brigade/bin/brigade-vacuum /usr/bin/brigade-vacuum
CMD ["/usr/bin/brigade-vacuum"]
