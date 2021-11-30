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

Note that selection of a viable third-party auth strategy is a prerequisite for
enabling the suite of user authorization commands.
## Background

The motivation behind this design stems from one of the core guiding principles
in Brigade 2: Provide an abstraction between users and the underlying substrate
(Kubernetes), such that users of Brigade do not need to interact with (or have
any presence on) the substrate itself. Therefore, the need to source user
identity from another authority became apparent. Rather than implement such a
system in Brigade itself, we decided to utilize well-known external services
to fulfill this need.

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
  ## Options for authenticating via a third-party authentication provider.
  thirdPartyAuth:
    ## Here we enter our chosen strategy of oidc
    strategy: oidc
    ## Here we inject the values from the OIDC provider
    oidc:
      providerURL: https://accounts.google.com
      clientID: <client ID>
      clientSecret: <client secret>
    ## User Session TTL dictates the default time-to-live for user sessions.
    ## Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    ## For example, "60s", "2h45m", "168h" (1 week)
    userSessionTTL: 168h
    ## If there are users that should be granted admin privileges in Brigade
    ## from the moment they authenticate, they should be listed here.  For
    ## instance, the operator doing the deployment and/or the user who will
    ## be in charge of assigning authorization permissions to authenticated
    ## users.
    ## Otherwise, permissions will need to be manually configured for each
    ## authenticated user.
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

[values.yaml]: https://github.com/brigadecore/brigade/blob/main/charts/brigade/values.yaml

### GitHub OAuth Provider

To set up GitHub authentication with Brigade, you'll create a [GitHub OAuth
App] with the authorization callback URL set to point to the
`/v2/session/auth` endpoint on Brigade's API server address, e.g. the hostname
reserved when setting up ingress or the external IP address if no ingress is
used.  You'll then use the GitHub OAuth Apps's client ID and a generated client
secret in the [values.yaml] file for the Brigade Helm chart.

  1. Follow the [GitHub OAuth App] creation instructions, supplying your choice
    of values for `Application Name`, `Homepage URL` and `Application description`.
  1. For the `Authorization callback URL`, you'll supply the a value based
    on the DNS hostname and path mentioned above.  For example, if the DNS
    hostname for the Brigade API server is `mybrigade.example.com`, the value
    would be:
      ```
      https://mybrigade.example.com/v2/session/auth
      ```
  1. Click 'Register application'
  1. On the App settings page, there is now a Client ID string. You'll use this
    value soon.
  1. Under `Client secrets`, click `Generate a new client secret`
  1. Save this client secret value now, as it won't be displayed again
  1. Click 'Update application'

Now that you have values for the GitHub OAuth App's client ID and client
secret, you're ready to update our Brigade chart values file with this auth
configuration.

All of it goes under the `apiserver.thirdPartyAuth` section, which we show
here, with other nearby sections omitted:

```yaml
apiserver:
  ## Options for authenticating via a third-party authentication provider.
  thirdPartyAuth:
    ## Here we enter our chosen strategy of github
    strategy: github
    ## Here we inject the values from the GitHub OAuth App
    github:
      clientID: <client ID>
      clientSecret: <client secret>
      ## If only users from specific GitHub organizations should be allowed
      ## to authenticate, list them here.  Otherwise, users from any GitHub
      ## organization may attempt to authenticate.
      allowedOrganizations:
    ## User Session TTL dictates the default time-to-live for user sessions.
    ## Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    ## For example, "60s", "2h45m", "168h" (1 week)
    userSessionTTL: 168h
    ## If there are users that should be granted admin privileges in Brigade
    ## from the moment they authenticate, they should be listed here.  For
    ## instance, the operator doing the deployment and/or the user who will
    ## be in charge of assigning authorization permissions to authenticated
    ## users.
    ## Otherwise, permissions will need to be manually configured for each
    ## authenticated user.
    admins:
      - <GitHub username>
```


[GitHub OAuth App]: https://docs.github.com/en/developers/apps/creating-an-oauth-app
[Configuring Github Authentication]: /topics/operators/deploy#configuring-github-authentication
