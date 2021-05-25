const { events, Job } = require("brigadier");

events.on("exec", exec);

function exec(e, p) {
    let j1 = new Job("j1", "alpine:3.7", ["echo hello"]);
    let j2 = new Job("j2", "alpine:3.7", ["echo goodbye"]);

    j1.run()
    .then(() => {
        return j2.run()
    })
    .then(() => {
        console.log("done");
    });
};