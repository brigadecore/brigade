events.push = function(e) {
  testName = "with-secrets"
  j = new Job(testName)
  j.secrets = {"SUPER_SECRET": "myConfigValue"}
  j.run()

  p = mockPods[testName]
  found = _.findWhere(p.spec.containers[0].env, {name: "SUPER_SECRET"})
  if (!found) {
    console.log(JSON.stringify(p.spec.containers[0].env))
    throw "Expected SUPER_SCRET"
  }
  if (found.valueFrom.secretKeyRef.name != e.projectId) {
    throw "project ID not used for secret name."
  }
  if (found.valueFrom.secretKeyRef.key != "myConfigValue") {
    throw "expected myConfigValue, got " + found.valueFrom.secretKeyRef.key
  }
}
