import { events, Job } from "@brigadecore/brigadier"
import { uniqueNamesGenerator, adjectives, animals } from "unique-names-generator"

function newJobName(): string {
  return uniqueNamesGenerator({
    dictionaries: [adjectives, animals],
    length: 2,
    separator: "-"
  })
}

events.on("brigade.sh/cli", "exec", async event => {
  const job1Name = newJobName()
  let job1 = new Job(job1Name, "debian:latest", event)
  job1.primaryContainer.command = ["echo"]
  job1.primaryContainer.arguments = [`Hello, ${job1Name}!`]

  const job2Name = newJobName()
  let job2 = new Job(job2Name, "alpine:latest", event)
  job2.primaryContainer.command = ["echo"]
  job2.primaryContainer.arguments = [`Hello, ${job2Name}!`]
  
  const job3Name = newJobName()
  let job3 = new Job(job3Name, "centos:latest", event)
  job3.primaryContainer.command = ["echo"]
  job3.primaryContainer.arguments = [`Hello, ${job3Name}!`]

  const job4Name = newJobName()
  let job4 = new Job(job4Name, "ubuntu:latest", event)
  job4.primaryContainer.command = ["echo"]
  job4.primaryContainer.arguments = [`Hello, ${job4Name}!`]

  await Job.concurrent(
    Job.sequence(job1, job2),
    Job.sequence(job3, job4)
  ).run()
})

events.process()
