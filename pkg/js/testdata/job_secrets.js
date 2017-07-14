events.push = function(e) {
  testName = "with-secrets"
  j = new Job(testName)
  j.env = {"SUPER_SECRET": project.secrets.dbPassword}
  j.run()

  p = mockPods[testName]
  found = _.findWhere(p.spec.containers[0].env, {name: "SUPER_SECRET"})
  if (!found) {
    console.log(JSON.stringify(p.spec.containers[0].env))
    throw "Expected SUPER_SECRET"
  }
  if (found.value != "mySecretPassword") {
    throw "expected myConfigValue, got " + found.value
  }
}
