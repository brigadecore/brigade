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

wg.run()