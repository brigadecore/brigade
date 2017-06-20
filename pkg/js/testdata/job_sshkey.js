events.push = function(e){

  e.repo.sshKey = "my-ssh-key"

  j = new Job("test-with-key")
  j.tasks = [
    "echo hello"
  ]
  j.run()

  p = mockPods["test-with-key"]

  found = _.findWhere(p.spec.containers[0].env, {name: "ACID_REPO_KEY", value: "my-ssh-key"})
  console.log(JSON.stringify(p.spec.containers[0].env))
  if (!found) {
    throw "Expected exactly one pod"
  }
}
