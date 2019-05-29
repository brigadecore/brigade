FROM quay.io/deis/lightweight-docker-go:v0.6.0
ARG LDFLAGS
ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/brigadecore/brigade
COPY brig/ brig/
COPY pkg/ pkg/
COPY vendor/ vendor/
RUN go build -ldflags "$LDFLAGS" -o bin/brig ./brig/cmd/brig

FROM alpine:3.8
RUN apk add --no-cache ca-certificates && update-ca-certificates
COPY --from=0 /go/src/github.com/brigadecore/brigade/bin/brig /usr/bin/brig
CMD ["/usr/bin/brig"]
