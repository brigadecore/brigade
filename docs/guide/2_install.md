# Installing Acid

_This part is a work-in-progress because Acid is still developer-oriented_

Acid is deployed via Helm. Here are the steps:

1. Make sure `helm` is installed, and `helm version` returns the correct server.
2. Add the Acid repo: `helm repo add https://deis.github.io/acid`
3. Install Acid: `helm install acid/acid --name acid-server`

At this point, you have a running Acid service. You can use `helm get acid-server` and other
Helm tools to examine your running Acid server.

## Cluster Ingress

By default, Acid is configured to set up a service as a load balancer for your Acid
build system. To find out your IP address, run:

```console
$ kubectl get svc acid-server-acid
NAME                  CLUSTER-IP    EXTERNAL-IP    PORT(S)          AGE
maudlin-quokka-acid   10.0.110.59   135.15.52.20   7744:31558/TCP   45d
```

(Note that `acid-server-acid` is just the name of the Helm release (`acid-server`) with `-acid` appended)

The `EXTERNAL-IP` field is the IP address that external services, such as GitHub,
will use to trigger actions.

Note that this is just one way of configuring Acid to receive inbound connections.
Acid itself does not care how traffic is routed to it. Those with operational knowledge
of Kubernetes may wish to use another method of ingress routing.
