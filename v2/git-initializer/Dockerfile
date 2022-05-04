FROM brigadecore/go-libgit2:v0.1.0 as builder

ARG VERSION
ARG COMMIT

WORKDIR /src
COPY sdk/ sdk/
WORKDIR /src/v2
COPY v2/go.mod go.mod
COPY v2/go.sum go.sum
RUN go mod download
COPY v2/git-initializer/ git-initializer/
COPY v2/internal/ internal/

# Despite CGO being involved, this builds a statically linked binary
RUN go build \
  -tags static,system_libgit2 \
  -o ../bin/git-initializer \
  -ldflags "-extldflags '-static -lgcrypt -lgpg-error' -w -X github.com/brigadecore/brigade-foundations/version.version=$VERSION -X github.com/brigadecore/brigade-foundations/version.commit=$COMMIT" \
  ./git-initializer

# Note: Cannot use gcr.io/distroless/static:nonroot because we still need glibc
FROM gcr.io/distroless/base:nonroot as final

COPY --from=builder /src/bin/ /brigade/bin/

ENTRYPOINT ["/brigade/bin/git-initializer"]
