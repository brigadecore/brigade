events.push = function(e) {
  // This should result in a default pod created with default image and no
  // commands.
  (new Job("empty")).run()
  p = mockPods["empty"]
  if (!p.spec.containers[0].image) {
    throw "expected an empty job to still have a pod with an image"
  }
}
