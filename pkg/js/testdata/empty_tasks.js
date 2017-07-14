events.push = function(e) {
  // This should result in a default pod created with default image and no
  // commands.
  (new Job("empty")).run()
  var p = mockPods["empty"]
  if (!p.spec.containers[0].image) {
    throw "expected an empty job to still have a pod with an image"
  }

  var sidecar = p.metadata.annotations["pod.beta.kubernetes.io/init-containers"]
  if (sidecar.indexOf("mySidecar") < 0) {
    throw "expected to find mySidecar in " + sidecar
  }

  var labels = p.metadata.labels
  if (labels.belongsto != "github.com-owner-repo") {
    throw "expected github.com-owner-repo, got " + labels.belongsto
  }
}
