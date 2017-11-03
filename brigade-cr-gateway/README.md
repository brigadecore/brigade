# Brigade Container Registry Gateway

This server provides a gateway for container registry webhooks.

Known-working container registries are:

- DockerHub
- Azure Container Registry (ACR)

If you do not see your preferred registry above, you can do any of the following:

1. Test this gateway and see if it works for you. If it does, please report to
  us.
2. Write a custom gateway. This may be necessary if your chosen container registry
  uses even moderately exotic auth patterns.
3. File an issue in the issue queue here, and see if you can rally some support
  for building one.
