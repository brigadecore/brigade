// The internals for running tasks. This must be loaded before any of the
// objects that use run().
//
// All Kubernetes API calls should be localized here. Other modules should not
// call 'kubernetes' directly.

var sidecarImage = "acidic.azurecr.io/vcs-sidecar:latest"

// wait takes a job-like object and waits until it succeeds or fails.
function waitForJob(job) {
  var my = job

  for (var i = 0; i < 300; i++) {
    console.log("checking status of " + my.podName)
    var k = kubernetes.withNS("default")
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
  runner.metadata.labels.belongsto = e.repo.name.replace("/", "-")
  runner.metadata.labels.commit = e.commit

  // Add env vars.
  var envVars = []

  _.each(job.env, function(val, key) {
    envVars.push({name: key, value: val});
  });

  // Add secrets as env vars.
  _.each(job.secrets, function(val, key) {

    // Some secrets we explicitly block.
    if (_.contains(["secret"], val)) {
      return
    }

    // Get secrets from the given secName
    envVars.push({
      name: key,
      valueFrom: {
        secretKeyRef: {name: e.projectId, key: val}
      }
    });
  });

  // Do we still want to add this to the image directly? While it is a security
  // thing, not adding it would result in users not being able to push anything
  // upstream into the pod.
  if (e.repo.sshKey) {
    envVars.push({
      name: "ACID_REPO_KEY",
      value: e.repo.sshKey
    })
  }

  // Add top-level env vars. These must override any attempt to set the values
  // to something else.
  envVars.push({ name: "CLONE_URL", value: e.repo.cloneURL })
  envVars.push({ name: "SSH_URL", value: e.repo.sshURL })
  envVars.push({ name: "GIT_URL", value: e.repo.gitURL })
  envVars.push({ name: "HEAD_COMMIT_ID", value: e.commit })
  runner.spec.containers[0].env = envVars

  // Add config map volume
  runner.spec.volumes = [
    { name: cmName, configMap: {name: cmName }},
    { name: "vcs-sidecar", emptyDir: {}}
    // , { name: "idrsa", secret: { secretName: secName }}
  ];
  runner.spec.containers[0].volumeMounts = [
    { name: cmName, mountPath: "/hook"},
    { name: "vcs-sidecar", mountPath: "/src"}
    // , { name: "idrsa", mountPath: "/hook/ssh", readOnly: true}
  ];

  // Add the sidecar.
  var sidecar = sidecarSpec(e, "/src", sidecarImage)

  // TODO: convert this to an init container with Kube 1.6
  // runner.spec.initContainers = [sidecar]
  runner.metadata.annotations["pod.beta.kubernetes.io/init-containers"] = "[" + JSON.stringify(sidecar) + "]"

  // Join the tasks to make a new command:
  // TODO: This should probably generate a shell script, starting with
  // something like set -eo pipefail, and using newlines instead of &&.
  var newCmd = job.tasks.join(" && ")

  cm.data["main.sh"] = newCmd

  var k = kubernetes.withNS("default")

  console.log("Creating configmap " + cm.metadata.name)
  //console.log(JSON.stringify(cm))
  k.extensions.configmap.create(cm)
  console.log("Creating pod " + runner.spec.containers[0].name)
  //console.log(JSON.stringify(runner))
  k.coreV1.pod.create(runner)
  console.log("running...")

  return runnerName;
}

function sidecarSpec(e, local, image) {
  var imageTag = image
  var repoURL = e.repo.cloneURL

  if (!imageTag) {
    imageTag = "acid/vcs-sidecar:latest"
  }

  if (e.repo.sshKey != "") {
    repoURL = e.repo.sshURL
  }

  spec = {
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

  if (e.repo.sshKey) {
   spec.env.push({
      name: "ACID_REPO_KEY",
      value: e.repo.sshKey
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
      "namespace": "default",
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
        "namespace": "default",
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
