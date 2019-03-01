const { events, Job } = require("brigadier");

events.on("exec", exec);

async function exec(e, p) {
    let j1 = new Job("j1", "alpine:3.7", ["echo hello"]);
    let j2 = new Job("j2", "alpine:3.7", ["echo goodbye"]);

    await j1.run();
    await j2.run();
    console.log("done");
}