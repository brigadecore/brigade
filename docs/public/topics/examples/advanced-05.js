const {events, Job, Group} = require("brigadier");

class MyJob extends Job {
  constructor(name) {
    super(name, "alpine:3.7");
    this.tasks = [
      "echo hello",
      "echo world"
    ];
  }
}

events.on("exec", (e, p) => {
  const j1 = new MyJob("j1")
  j1.tasks.push("echo goodbye");
  
  const j2 = new MyJob("j2")

  Group.runEach([j1, j2])
});