// This is the runner wrapping script.
console.log("GOT HERE")
console.log(pushRecord.repository.name)

acidImage = "acidrun:latest"

/*
TODO: define a prototype for a job
 */

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

  if (job.image) {
    runner.spec.containers[0].image = job.image
  }

  // Join the tasks to make a new command:
  newCmd = job.tasks.join(" && ")
  cm.data["main.sh"] = newCmd

  k = kubernetes.withNS("default")
  console.log("Creating configmap " + cm.metadata.name)
  console.log(JSON.stringify(cm))
  k.extensions.configmap.create(cm)
  console.log("Creating pod " + runner.spec.containers[0].name)
  console.log(JSON.stringify(runner))
  k.coreV1.pod.create(runner)
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
