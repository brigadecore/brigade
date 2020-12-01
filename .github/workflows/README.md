# A Note on Our Use of GitHub Actions

Nearly all of Brigade's own pipelines are implemented using Brigade. (i.e. We
eat our own dog food.)

Brigade Workers and Jobs are implemented as Kubernetes pods comprised of one or
more OCI containers. Building any of Brigade's many Linux-based OCI images is a
task that can be accomplished by a Brigade Job as the process of building a new
Linux-based OCI image _inside_ a Linux-based OCI container is a well-understood
and well-documented pattern colloquially called "Docker in Docker" or simply
"DinD."

Building a Windows-based OCI image inside a Windows-based OCI container is
_not_, as of this writing, a common, well-understood, or well-documented
pattern. Building Brigade's one Windows-based OCI image, therefore, requires the
use of a physical Windows machine or virtual machine having an appropriate
kernel. This requirement, unfortunately, eliminates the possibility of building
such an image using Brigade itself.

For this reason, and this reason only, the maintainers have opted (for now) to
utilize GitHub Actions _only_ for building this one Windows-based OCI image. It
is expected that this decision may be revisited as the state of the art in
building Windows-based OCI images advances.

Further, building Windows-based OCI images is time consuming in comparison to
building Linux-based OCI images. In light of this, the maintainers have elected
_not_ to unconditionally require the build of Brigade's one Windows-based OCI
image to complete as a prerequisite for merging a PR. Maintainers _will_ require
that build to pass in the event that the PR in question modifies the
corresponding Dockerfile.
