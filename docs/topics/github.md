# GitHub Integration

Acid provides GitHub integration for triggering Acid builds from GitHub events.

Acid integrates with GitHub by providing GitHub webhook implementations for the `push`
and `pull_request` events. You must be running `acid-gateway` in a way that makes
it available to GitHub. (For example, assign it a publicly routable IP and domain name.)

## Configuring

To add an Acid project to GitHub:

1. Go to "Settings"
2. Click "Webhooks"
3. Click the "Add webhook" button
4. For "Payload URL", add the URL: "http://YOUR_HOSTNAME:7744/events/github"
5. For "Content type", choose "application/json"
6. For "Secret", use the secret you configured in your Helm config.
7. Choose "Just the push event" or choose "push" and "pull_request".

![GitHub Webhook Config](../intro/img/img4.png)

You may use GitHub's testing page to verify that GitHub can successfully send an event to
the Acid gateway.
