# GitHub Integration

Brigade provides GitHub integration for triggering Brigade builds from GitHub events.

Brigade integrates with GitHub by providing GitHub webhook implementations for the `push`
and `pull_request` events. You must be running `brigade-gateway` in a way that makes
it available to GitHub. (For example, assign it a publicly routable IP and domain name.)

## Configuring

To add an Brigade project to GitHub:

1. Go to "Settings"
2. Click "Webhooks"
3. Click the "Add webhook" button
4. For "Payload URL", add the URL: "http://YOUR_HOSTNAME:7744/events/github"
5. For "Content type", choose "application/json"
6. For "Secret", use the secret you configured in your Helm config.
7. Choose "Just the push event" or choose "push" and "pull_request".

![GitHub Webhook Config](../intro/img/img4.png)

You may use GitHub's testing page to verify that GitHub can successfully send an event to
the Brigade gateway.

## Connecting to Private GitHub Repositories (or Using SSH)

Sometimes it is better to configure Brigade to interact with GitHub via SSH. For example, if
your repository is private and you don't want to allow anonymous Git clones, you may need
to use an SSH GitHub URL. To use GitHub with SSH, you will also need to create a
Deployment Key.

To create a new GitHub Deployment Key, generate an SSH key. On UNIX-like systems, this is
done with `ssh-keygen -f ./github_deployment_key`. When prompted to set a passphrase, _do not set a passphrase_.

```console
ssh-keygen -f ./github_deployment_key
Generating public/private rsa key pair.
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
Your identification has been saved in ./github_deployment_key.
Your public key has been saved in ./github_deployment_key.pub.
...
```
In GitHub, navigate to your project, choose *Settings* (the gear icon), then choose
*Depoyment Keys* from the left-hand navigation. Click the *Add deploy key* button.

The *Title* field should be something like `brigade-checkout`, though the intent of this
field is just to help you remember that this key was used by Brigade.

The *Key* field should be the content of the `./github_deployment_key.pub` file generated
by `ssh-keygen` above.

Save that key.

Inside of your project configuration for your `brigade-project`, make sure to add your key:

myvalues.yaml:
```
project: "my/brigade-project"
repository: "github.com/my/brigade-project"
# This is an SSH clone URL
cloneURL: "git@github.com:my/brigade-project.git"
# paste your entire key here:
sshKey: |-
  -----BEGIN RSA PRIVATE KEY-----
  MIIEowIBAAKCAQEAupolYH/x2+V+L15ci3PU75GX8aKTWZzCPkX3qNqRqiO5q0LV
  nMIVeMSqrLDHSGnbUF6DN3EgKuwdv0bfiq3Cz1rjtszQX6ti50ICObGphU+6dTwO
  # removed some lines
  9KjBbQKBgA23dOOF98EjLcCZm/lky+Ifu2ZSbi+5N8MlbP3+5rWIgw74iAo6KHFb
  v/mHCUT7SWguIdNGzdAD+wYHG2W14fu+IQCWQ6oaZauHHqlxGrXH
  -----END RSA PRIVATE KEY-----
# The rest of your config...
```

Then you can install (or upgrade) your project:

```
$ helm install -n my-project brigade/brigade-project -f myvalues.yaml
```

Now your project is configured to clone via SSH using the deployment key we generated.
