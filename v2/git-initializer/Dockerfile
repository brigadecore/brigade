FROM --platform=$BUILDPLATFORM brigadecore/go-tools:v0.7.0 as builder

ARG VERSION
ARG COMMIT
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0

WORKDIR /src
COPY sdk/ sdk/
WORKDIR /src/v2
COPY v2/git-initializer/ git-initializer/
COPY v2/internal/ internal/
COPY v2/go.mod go.mod
COPY v2/go.sum go.sum
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
  -o ../bin/git-initializer \
  -ldflags "-w -X github.com/brigadecore/brigade-foundations/version.version=$VERSION -X github.com/brigadecore/brigade-foundations/version.commit=$COMMIT" \
  ./git-initializer

FROM gcr.io/distroless/static:nonroot as final

COPY --from=builder /src/bin/ /brigade/bin/

ENTRYPOINT ["/brigade/bin/git-initializer"]
