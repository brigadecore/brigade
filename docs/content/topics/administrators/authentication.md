---
title: Authentication
description: Authentication configuration for Brigade
section: administrators
weight: 1
aliases:
  - /authentication
  - /topics/authentication.md
  - /topics/administrators/authentication.md
---

# Brigade Authentication

Brigade utilizes third-party providers for its user authentication strategy.
We'll go over the available options and their configuration in this document.

Note that Brigade's default mode of root user access precludes functionality
related to user management or authorization. Therefore, selection of a viable
third-party auth strategy (and thus disabling root user mode) is a
prerequisite for configuring user authorization.

## Third-Party Authentication Options

Currently, there are two authentication options for integration with Brigade.

  * An [OpenID Connect] provider. Example services include [Google
  Identity Platform] and [Azure Active Directory].
  * [GitHub's OAuth Provider]

Additionally, the Brigade server must be running with TLS enabled when using
third-party authentication.  See the [Deploy] doc for more details on securing
your Brigade deployment.

[OpenID Connect]: https://openid.net/connect/
[Google Identity Platform]: https://cloud.google.com/identity-platform
[Azure Active Directory]: https://azure.microsoft.com/en-us/services/active-directory/
[GitHub's OAuth Provider]: https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps
[Deploy]: /topics/operator/deploy

### OpenID Connect Provider

To use an OpenID Connect provider for authentication into Brigade, you'll need
the following values after choosing and configuring your preferred provider:

  * Provider URL, e.g. `https://accounts.google.com` or
    `https://login.microsoftonline.com/<Azure Tenant ID>/v2.0`
  * Client ID
  * Client Secret

These values can then be provided in the [values.yaml] file for the Brigade
Helm chart. They'll go under the `apiserver.thirdPartyAuth` section.  Here is
an example:

```yaml
apiserver:
  rootUser:
    enabled: false
  thirdPartyAuth:
    strategy: oidc
    oidc:
      providerURL: https://accounts.google.com
      clientID: <client ID>
      clientSecret: <client secret>
    userSessionTTL: 168h
    admins:
      - myemail@example.com
```

Note that `strategy` must be set to `oidc`. The default time-to-live for user
sessions (retained in the example above) is 1 week/168 hours. Lastly, the
`admins` field is an optional list of email addresses for users that should
be granted full admin privileges upon first login to Brigade.

Note: since these privileges are granted on first login, adding/revoking these
permissions for users who have already logged in for the first time must be
done via the brig CLI directly rather than re-deploying with differing
configuration.

[values.yaml]: https://github.com/brigadecore/brigade/blob/v2/charts/brigade/values.yaml

### GitHub OAuth Provider

To use [GitHub's Oauth Provider], please see the
[Configuring GitHub Authentication] section of the [Deploy] doc for
setup and configuration.

[Configuring Github Authentication]: /topics/operators/deploy#configuring-github-authentication
