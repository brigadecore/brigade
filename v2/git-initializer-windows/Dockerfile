FROM golang:1.18.1-windowsservercore-1809 as builder

ARG VERSION
ARG COMMIT
ENV CGO_ENABLED=0

WORKDIR /src
COPY sdk/ sdk/
WORKDIR /src/v2
COPY v2/git-initializer/ git-initializer/
COPY v2/internal/ internal/
COPY v2/go.mod go.mod
COPY v2/go.sum go.sum

RUN go build \
  -o ../bin/git-initializer.exe \
  -ldflags \"-w -X github.com/brigadecore/brigade-foundations/version.version=$env:VERSION -X github.com/brigadecore/brigade-foundations/version.commit=$env:COMMIT\" \
  ./git-initializer

FROM mcr.microsoft.com/windows/nanoserver:1809

COPY --chown=ContainerUser:ContainerUser v2/git-initializer/ssh_config /Users/ContainerUser/.ssh/config
COPY --from=builder /git /git/
COPY --from=builder /src/bin/ /brigade/bin/

USER ContainerAdministrator
RUN setx /M PATH "%PATH%;C:\git\cmd"
USER ContainerUser

ENTRYPOINT ["/brigade/bin/git-initializer.exe"]
