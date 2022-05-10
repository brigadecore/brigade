FROM --platform=$BUILDPLATFORM brigadecore/go-tools:v0.9.0 as builder

ARG VERSION
ARG COMMIT
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0

WORKDIR /src
COPY sdk/ sdk/
WORKDIR /src/v2
COPY v2/go.mod go.mod
COPY v2/go.sum go.sum
RUN go mod download
COPY v2/git-initializer/ git-initializer/
COPY v2/internal/ internal/

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
  -o ../bin/git-initializer \
  -ldflags "-w -X github.com/brigadecore/brigade-foundations/version.version=$VERSION -X github.com/brigadecore/brigade-foundations/version.commit=$COMMIT" \
  ./git-initializer

FROM alpine:3.15.4 as final

RUN apk update \
    && apk add git openssh-client \
    && addgroup -S -g 65532 nonroot \
    && adduser -S -D -u 65532 -g nonroot -G nonroot nonroot

COPY --chown=nonroot:nonroot v2/git-initializer/ssh_config /home/nonroot/.ssh/config
COPY --from=builder /src/bin/ /brigade/bin/

USER nonroot

ENTRYPOINT ["/brigade/bin/git-initializer"]
