const { events, Job } = require("@brigadecore/brigadier")
// TODO: require published version of this project
// Or bring in locally?

const projectName = "brigadier"

events.on("exec", (e, p) => {
    return build(e, p).run();
})

events.on("check_suite:requested", runSuite)
events.on("check_suite:rerequested", runSuite)
events.on("check_run:rerequested", runCheck);
events.on("issue_comment:created", handleIssueComment);
events.on("issue_comment:edited", handleIssueComment);

function build(e, project) {
    var build = new Job(`${projectName}-build`, "node:12.3.1-stretch");

    build.tasks = [
        "cd /src",
        "yarn install",
        "yarn compile",
        "yarn test",
        "yarn audit"
    ];

    return build;
}

// Check represents a simple Check run,
// consisting of a name and an action (javascript function)
class Check {
    constructor(name, action) {
        this.name = name;
        this.action = action;
    }
}

// Checks represent a list of Checks that by default are run in the form
// of a check suite, but may be run individually
Checks = {
    "build": new Check("build", build)
};

// runCheck is the default function invoked on a check_run:* event
//
// It determines which check is being requested (from the payload body)
// and runs this particular check, or else throws an error if the check
// is not found
function runCheck(e, p) {
    payload = JSON.parse(e.payload);

    name = payload.body.check_run.name;
    check = Checks[name];

    if (typeof check !== 'undefined') {
        checkRun(e, p, check);
    } else {
        throw new Error(`No check found with name: ${name}`);
    }
}

// checkRun is a GitHub Check Run
//
// It runs the provided check, wrapped in notification jobs
// to update GitHub along the way
function checkRun(e, p, check) {
    console.log(`Check requested: ${check.name}`);

    // Create Notification object (which is a Job to update GH using the Checks API)
    var note = new Notification(`${check.name}`, e, p);
    note.conclusion = "";
    note.title = `Run ${check.name}`;
    note.summary = `Running ${check.name} for ${e.revision.commit}`;
    note.text = `Ensuring ${check.name} passes.`

    // Send notification, then run, then send pass/fail notification
    return notificationWrap(check.action(e, p), note)
}

// Our Check Run Suite is composed of GitHub Check Runs,
// which will run in parallel and report their results independently to GitHub
function runSuite(e, p) {
    var checkRuns = new Array();
    for (check of Object.values(Checks)) {
        checkRuns.push(checkRun(e, p, check));
    }

    // Important: To prevent Promise.all() from failing fast, we catch and
    // return all errors. This ensures Promise.all() always resolves. We then
    // iterate over all resolved values looking for errors. If we find one, we
    // throw it so the whole build will fail.
    //
    // Ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/all#Promise.all_fail-fast_behaviour
    return Promise.all(checkRuns)
        .then((values) => {
            values.forEach((value) => {
                if (value instanceof Error) throw value;
            });
        });
}

// handleIssueComment handles an issue_comment event, parsing the comment text
// and determining whether or not to trigger an action
function handleIssueComment(e, p) {
    console.log("handling issue comment....")
    payload = JSON.parse(e.payload);

    // Extract the comment body and trim whitespace
    comment = payload.body.comment.body.trim();

    // Here we determine if a comment should provoke an action
    switch (comment) {
        // Currently, the do-all '/brig run' comment is supported,
        // for (re-)triggering the default Checks suite
        case "/brig run":
            return runSuite(e, p);
        default:
            console.log(`No applicable action found for comment: ${comment}`);
    }
}

// Helpers
// TODO: use/consume corresponding utils in this project for this (and other?) logic

// A GitHub Check Suite notification
class Notification {
    constructor(name, e, p) {
        this.proj = p;
        this.payload = e.payload;
        this.name = name;
        this.externalID = e.buildID;
        this.detailsURL = `https://brigadecore.github.io/kashti/builds/${e.buildID}`;
        this.title = "running check";
        this.text = "";
        this.summary = "";

        // count allows us to send the notification multiple times, with a distinct pod name
        // each time.
        this.count = 0;

        // One of: "success", "failure", "neutral", "cancelled", or "timed_out".
        this.conclusion = "neutral";
    }

    // Send a new notification, and return a Promise<result>.
    run() {
        this.count++
        var j = new Job(`${this.name}-${this.count}`, "brigadecore/brigade-github-check-run:latest");
        j.imageForcePull = true;
        j.env = {
            CHECK_CONCLUSION: this.conclusion,
            CHECK_NAME: this.name,
            CHECK_TITLE: this.title,
            CHECK_PAYLOAD: this.payload,
            CHECK_SUMMARY: this.summary,
            CHECK_TEXT: this.text,
            CHECK_DETAILS_URL: this.detailsURL,
            CHECK_EXTERNAL_ID: this.externalID
        }
        return j.run();
    }
}

// Helper to wrap a job execution between two notifications.
async function notificationWrap(job, note, conclusion) {
    if (conclusion == null) {
        conclusion = "success"
    }
    await note.run();
    try {
        let res = await job.run()
        const logs = await job.logs();

        note.conclusion = conclusion;
        note.summary = `Task "${job.name}" passed`;
        note.text = "```" + res.toString() + "```\nTest Complete";
        return await note.run();
    } catch (e) {
        const logs = await job.logs();
        note.conclusion = "failure";
        note.summary = `Task "${job.name}" failed for ${e.buildID}`;
        note.text = "```" + logs + "```\nFailed with error: " + e.toString();
        try {
            return await note.run();
        } catch (e2) {
            console.error("failed to send notification: " + e2.toString());
            console.error("original error: " + e.toString());
            return e2;
        }
    }
}
