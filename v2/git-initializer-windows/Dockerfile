# escape=`

FROM brigadecore/win-go-tools:v0.2.0 as builder

# As of this writing, pacman installs a newer libgit2 than we are able to use.
# Here, we download and install an older version.
RUN curl `
      -L `
      https://repo.msys2.org/mingw/mingw64/mingw-w64-x86_64-libgit2-1.2.0-3-any.pkg.tar.zst `
      -o C:\windows\temp/libgit2.pkg.tar.zst `
    && bash -l -c "pacman --noconfirm -U /c/windows/temp/libgit2.pkg.tar.zst"

WORKDIR C:\\msys64\\brigade
COPY sdk\\ sdk\\
WORKDIR C:\\msys64\\brigade\\v2
COPY v2\\go.mod go.mod
COPY v2\\go.sum go.sum
RUN bash -l -c " `
      cd /brigade/v2 `
      && PATH=$PATH:/mingw64/bin `
      GOROOT=/mingw64/lib/go `
      go mod download `
    "
COPY v2\\git-initializer\\ git-initializer\\
COPY v2\\internal\\ internal\\

# The `-tags static,system_libgit2` specified below instruct git2go on how to
# locate libgit2. It does NOT imply our binary is statically linked -- in fact,
# it is not. In theory, that can be accomplished by including
# `-extldflags 'static'` in the `-ldflags`, but we've had no luck getting that
# to work.
RUN bash -l -c " `
      cd /brigade/v2 `
      && PATH=$PATH:/mingw64/bin `
      GOROOT=/mingw64/lib/go `
      go build `
        -tags static,system_libgit2 `
        -ldflags \"-w -X github.com/brigadecore/brigade-foundations/version.version=$VERSION -X github.com/brigadecore/brigade-foundations/version.commit=$COMMIT\" `
        -o ../bin/git-initializer.exe `
        ./git-initializer `
    "

ENTRYPOINT [ "C:\\msys64\\brigade\\bin\\git-initializer.exe" ]

FROM mcr.microsoft.com/windows/nanoserver:1809 AS final

USER ContainerAdministrator

# Note that because we were unable to produce a statically-linked binary, we
# depend on a number of .dlls from the builder image, so we just copy them
# all over and add them to the path.
COPY --from=builder C:\\msys64\\mingw64\\bin\\ C:\\mingw64\\bin\\
RUN setx /M PATH "%PATH%;C:\mingw64\bin"

COPY --from=builder C:\\msys64\\brigade\\bin\\ C:\\brigade\\bin\\

USER ContainerUser

ENTRYPOINT ["C:\\brigade\\bin\\git-initializer.exe"]
