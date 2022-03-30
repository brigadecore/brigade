FROM --platform=$BUILDPLATFORM brigadecore/go-tools:v0.9.0 as builder

ARG VERSION
ARG COMMIT
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0

WORKDIR /src
COPY sdk/ sdk/
WORKDIR /src/v2
COPY v2/observer/ observer/
COPY v2/internal/ internal/
COPY v2/go.mod go.mod
COPY v2/go.sum go.sum
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
  -o ../bin/observer \
  -ldflags "-w -X github.com/brigadecore/brigade-foundations/version.version=$VERSION -X github.com/brigadecore/brigade-foundations/version.commit=$COMMIT" \
  ./observer

FROM gcr.io/distroless/static:nonroot as final

COPY --from=builder /src/bin/ /brigade/bin/

ENTRYPOINT ["/brigade/bin/observer"]
