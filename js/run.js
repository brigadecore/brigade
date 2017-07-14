// The internals for running tasks. This must be loaded before any of the
// objects that use run().
//
// All Kubernetes API calls should be localized here. Other modules should not
// call 'kubernetes' directly.

// wait takes a job-like object and waits until it succeeds or fails.
function waitForJob(job, e) {
  var my = job

  for (var i = 0; i < 300; i++) {
    console.log("checking status of " + my.podName)
    var k = kubernetes.withNS(project.kubernetes.namespace)
    var mypod = k.coreV1.pod.get(my.podName)

    console.log(JSON.stringify(mypod))
    console.log("Pod " + my.podName + " is in state " + mypod.status.phase)

    if (mypod.status.phase == "Failed") {
      throw "Pod " + my.podName + " failed to run to completion";
    }
    if (mypod.status.phase == "Succeeded") {
      return true
    }
    // Sleep for a defined amount of time.
    sleep(3)
  }
  throw "timed out waiting for pod " + my.podName + " to run"
}

// run runs a job for an event (e). It does not wait for the job to complete.
// This is a low-level primitive.
function run(job, e) {
  // $JOB-$TIME-$GITSHA
  var k8sName = job.name + "-" + Date.now() + "-" + e.commit.substring(0, 8);
  var cmName = k8sName
  var runnerName = k8sName
  var cm = newCM(cmName)
  var runner = newRunnerPod(runnerName, job.image)

  runner.metadata.labels.jobname = job.name
  runner.metadata.labels.belongsto = project.repo.name.replace(/\//g, "-")
  runner.metadata.labels.commit = e.commit

  // Add env vars.
  var envVars = []

  _.each(job.env, function(val, key) {
    envVars.push({name: key, value: val});
  });

  // Do we still want to add this to the image directly? While it is a security
  // thing, not adding it would result in users not being able to push anything
  // upstream into the pod.
  if (project.repo.sshKey) {
    envVars.push({
      name: "ACID_REPO_KEY",
      value: project.repo.sshKey
    })
  }

  // Add top-level env vars. These must override any attempt to set the values
  // to something else.
  envVars.push({ name: "CLONE_URL", value: project.repo.cloneURL })
  envVars.push({ name: "SSH_URL", value: project.repo.sshURL })
  envVars.push({ name: "GIT_URL", value: project.repo.gitURL })
  envVars.push({ name: "HEAD_COMMIT_ID", value: e.commit })
  envVars.push({ name: "CI", value: "true" })
  runner.spec.containers[0].env = envVars

  var mountPath = job.mountPath || "/src"

  // Add config map volume
  runner.spec.volumes = [
    { name: cmName, configMap: {name: cmName }},
    { name: "vcs-sidecar", emptyDir: {}}
    // , { name: "idrsa", secret: { secretName: secName }}
  ];
  runner.spec.containers[0].volumeMounts = [
    { name: cmName, mountPath: "/hook"},
    { name: "vcs-sidecar", mountPath: mountPath}
    // , { name: "idrsa", mountPath: "/hook/ssh", readOnly: true}
  ];

  // Add the sidecar.
  var sidecar = sidecarSpec(e, "/src", project.kubernetes.vcsSidecar)

  // TODO: convert this to an init container with Kube 1.6
  // runner.spec.initContainers = [sidecar]
  runner.metadata.annotations["pod.beta.kubernetes.io/init-containers"] = "[" + JSON.stringify(sidecar) + "]"

  var newCmd = "#!" + job.shell + "\n\n"

  // if shells that support the `set` command are selected, let's add some sane defaults
  if (job.shell == "/bin/sh" || job.shell == "/bin/bash") {
    newCmd += "set -e\n\n"
  }

  // Join the tasks to make a new command:
  if (job.tasks) {
    newCmd += job.tasks.join("\n")
  }

  cm.data["main.sh"] = newCmd
  console.log("using main.sh:\n" + cm.data["main.sh"])

  var k = kubernetes.withNS(project.kubernetes.namespace)

  console.log("Creating configmap " + cm.metadata.name)
  k.extensions.configmap.create(cm)
  console.log("Creating pod " + runner.spec.containers[0].name)
  k.coreV1.pod.create(runner)
  console.log("running...")

  return runnerName;
}

function sidecarSpec(e, local, image) {
  var imageTag = image
  var repoURL = project.repo.cloneURL

  if (!imageTag) {
    imageTag = "acid/vcs-sidecar:latest"
  }

  if (project.repo.sshKey != "") {
    repoURL = project.repo.sshURL
  }

  var spec = {
    name: "acid-vcs-sidecar",
    env: [
      { name: "VCS_REPO", value: repoURL },
      { name: "VCS_LOCAL_PATH", value: local },
      { name: "VCS_REVISION", value: e.commit },
    ],
    image: imageTag,
    command: ["/vcs-sidecar"],
    imagePullPolicy: "Always",
    volumeMounts: [
      { name: "vcs-sidecar", mountPath: local},
    ]
  }

  if (project.repo.sshKey) {
   spec.env.push({
      name: "ACID_REPO_KEY",
      value: project.repo.sshKey
    })
  }

  return spec
}

function newRunnerPod(podname, acidImage) {
  return {
    "kind": "Pod",
    "apiVersion": "v1",
    "metadata": {
      "name": podname,
      "labels": {
        "heritage": "Quokka",
        "managedBy": "acid"
      },
      "annotations": {}
    },
    "spec": {
      "restartPolicy": "Never",
      "containers": [
        {
          "name": "acidrun",
          "image": acidImage,
          "command": [
            "/bin/sh",
            "/hook/main.sh"
          ],
          // FIXME: Change to "IfNotPresent"
          "imagePullPolicy": "Always"
        }
      ]
    }
  };
}


function newCM(name) {
  return {
    "kind": "ConfigMap",
    "apiVersion": "v1",
    "metadata": {
        "name": name,
        "labels": {
            "heritage": "Quokka",
        },
    },
    "data": {
        "main.sh": "echo hello && echo goodbye"
    },
  };
}

exports.run = run
exports.waitForJob = waitForJob
