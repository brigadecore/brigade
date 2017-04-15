// This is the runner wrapping script.
console.log("GOT HERE")
console.log(pushRecord.repository.name)

// The default image is stock ubuntu 16.04 + make and git.
acidImage = "acid-ubuntu:latest"

// Prototype for Job.
function Job(name, tasks) {
  // Name will become the prefix for the pod/configmap names.
  this.name = name;
  // Tasks is the list of tasks to run. They are executed in sequence inside of
  // a shell (/bin/sh).
  this.tasks = tasks;

  // A collection of name/value pairs of environment variables.
  this.env = {};

  // The image and an optional tag.
  this.image = acidImage;

  // A map of ENV_VAR names and Secret names. This will populate the environment
  // variable with the value found in the secret.
  // This will override a matching env var from the env map.
  this.secrets = {}

  // run sends this job to Kubernetes.
  this.run = function(pushRecord) { run(this, pushRecord); };
}

function run(job, pushRecord) {
  cid = pushRecord.head_commit.id
  cmName = job.name + "-" + cid
  runnerName = job.name + "-" + cid

  cm = newCM(cmName)
  runner = newRunnerPod(runnerName)

  // Add env vars.
  envVars = []
  _.each(job.env, function(val, key, l) {
    envVars.push({name: key, value: val});
  });
  // Add secrets as env vars.
  _.each(job.secrets, function(val, key, l) {
    parts = val.split(".", 2)
    envVars.push({
      name: "key",
      valueFrom: {
        secretKeyRef: {name: parts[0], key: parts[1]}
      }
    })
  });
  // Add top-level env vars. These must override any attempt to set the values
  // to something else.
  envVars.push({ name: "CLONE_URL", value: pushRecord.repository.clone_url })
  envVars.push({ name: "HEAD_COMMIT_ID", value: pushRecord.head_commit.id })
  runner.spec.containers[0].env = envVars

  // Add config map volume
  runner.spec.volumes = [
    { name: cmName, configMap: {name: cmName }}
  ];
  runner.spec.containers[0].volumeMounts = [
    { name: cmName, mountPath: "/hook/data"}
  ];

  // Override the image only if the user sets one.
  if (job.image) {
    runner.spec.containers[0].image = job.image
  }

  // Join the tasks to make a new command:
  newCmd = job.tasks.join(" && ")
  cm.data["main.sh"] = newCmd

  k = kubernetes.withNS("default")
  console.log("Creating configmap " + cm.metadata.name)
  console.log(JSON.stringify(cm))
  //k.extensions.configmap.create(cm)
  console.log("Creating pod " + runner.spec.containers[0].name)
  console.log(JSON.stringify(runner))
  //k.coreV1.pod.create(runner)
  console.log("running...")
}

function newRunnerPod(podname) {
  return {
    "kind": "Pod",
    "apiVersion": "v1",
    "metadata": {
        "name": podname,
        "namespace": "default",
        "labels": {
          "heritage": "Quokka",
          "managedBy": "acid"
        }
    },
    "spec": {
        "containers": [
            {
                "name": "acidrun",
                "image": acidImage,
                "command": [
                    "/hook.sh"
                ],
                "imagePullPolicy": "IfNotPresent"
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

console.log("Loaded runner")
