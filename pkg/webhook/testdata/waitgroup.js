events.github.push = function (e) {
    console.log("Starting waitgroup")
    one = new Job("one")
    one.tasks = [
      "echo hello"
    ]

    two = new Job("two")
    two.tasks = [
      "echo world"
    ]

    wg = new WaitGroup()
    wg.add(one)
    wg.add(two)

    console.log("about to run")
    wg.run()
}
