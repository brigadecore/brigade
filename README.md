# Brigade 2: Event-based Scripting for Kubernetes

![build](https://badgr.brigade2.io/v1/github/checks/brigadecore/brigade/badge.svg?appID=99005&branch=v2)
[![slack](https://img.shields.io/badge/slack-brigade-brightgreen.svg?logo=slack)](https://kubernetes.slack.com/messages/C87MF1RFD)

<img width="100" align="left" src="logo.png">

Brigade 2 is currently in a _beta_ state and remains under active development,
but that effort is primarily oriented around improving the platform's stability.
Breaking changes to APIs are not anticipated.

<br clear="left"/>

## Introducing Brigade 2

Brigade 2 has been lovingly re-engineered from the ground up. We believe we've
remained faithful to the original vision of Brigade 1.x, and as such, much
general knowledge of Brigade 1.x can be carried over.

_But we've also learned a lot from Brigade 1.x._ Brigade 2 has been designed,
_explicitly_ to reduce the degree of Kubernetes knowledge required for success.
While Brigade 1.x was hailed as "event driven scripting for Kubernetes," Brigade
2 is "event driven scripting (for Kubernetes)." Moreover, great care has been
taken to improve security and scalability, and with our all new API and
complementary SDKs, we're also lowering barriers to integration.

We hope you'll enjoy this product refresh as much as we are.

## Getting Started

While comprehensive documentation remains a work in progress, our [quickstart]
is well-polished and we refer users wishing to test drive Brigade 2 to
[that documentation][quickstart].

[quickstart]: https://v2--brigade-docs.netlify.app/intro/quickstart/

## The Brigade 2 Ecosystem

Brigade 2's all new API lowers the bar to creating all manner of peripherals--
tooling, event gateways, and more. Even though we're still in beta, some great
integrations already exist!

### Gateways

Gateways receive events from upstream systems (the "outside world") and convert
them to Brigade events that are emitted into Brigade's event bus.

* [ACR (Azure Container Registry) Gateway](https://github.com/brigadecore/brigade-acr-gateway)
* [Bitbucket Gateway](https://github.com/brigadecore/brigade-bitbucket-gateway/tree/v2)
* [CloudEvents Gateway](https://github.com/brigadecore/brigade-cloudevents-gateway)
* [Docker Hub Gateway](https://github.com/brigadecore/brigade-dockerhub-gateway)
* [GitHub Gateway](https://github.com/brigadecore/brigade-github-gateway)
* [Slack Gateway](https://github.com/brigadecore/brigade-slack-gateway)

The [Brigade GitHub Gateway](https://github.com/brigadecore/brigade-github-gateway),
in particular, is already utilized extensively by the Brigade team in
combination with Brigade 2 to implement the project's own CI/CD. This is our
favorite gateway at the moment because it also reports event statuses upstream
to GitHub, making it a true showcase for what can be done with gateways.

### Monitoring

[Brigade Metrics](https://github.com/brigadecore/brigade-metrics) is a great way
to obtain operational insights into a Brigade 2 installation.

### Chaos Engineering

The Brigade team is utilizing
[Brigade Noisy Neighbor](https://github.com/brigadecore/brigade-noisy-neighbor)
to keep our own internal Brigade 2 installation under a steady load. We hope the
larger event volumes than what we generate on our own will help us to identify
and resolve bugs sooner.

### SDKs

Use any of these to develop your own integrations!

* [Brigade SDK for Go](https://github.com/brigadecore/brigade/tree/v2/sdk) (used by Brigade itself)
* [Brigade SDK for JavaScript](https://github.com/krancour/brigade-sdk-for-js) (and TypeScript)]
* [Brigade SDK for Rust](https://github.com/brigadecore/brigade-sdk-for-rust) (still a work-in-progress)

## Contributing

The Brigade project accepts contributions via GitHub pull requests. The
[Contributing](CONTRIBUTING.md) document outlines the process to help get your
contribution accepted.

## Support & Feedback

We have a slack channel!
[Kubernetes/#brigade](https://kubernetes.slack.com/messages/C87MF1RFD) Feel free
to join for any support questions or feedback, we are happy to help. To report
an issue or to request a feature open an issue
[here](https://github.com/brigadecore/brigade/issues)

## Code of Conduct

Participation in the Brigade project is governed by the
[CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).
