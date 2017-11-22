import "mocha"
import {assert} from "chai"
import * as mock from "./mock"

import * as k8s from "../src/k8s"
import {BrigadeEvent, Project} from "../src/events"
import {Job, Result, brigadeCachePath, brigadeStoragePath} from "../src/job"

import * as kubernetes from '@kubernetes/typescript-node'

describe("k8s", function() {

  describe("b64enc", () => {
    it('encodes the string "hello"', function() {
      assert.equal(k8s.b64enc("hello"), "aGVsbG8=")
    })
  })

  describe("b64dec", () => {
    it('decodes the string "aGVsbG8="', function() {
      assert.equal(k8s.b64dec("aGVsbG8="), "hello")
    })
  })

  describe("secretToProject", function() {
    it("converts secret to project", function() {
      let s = mockSecret()
      let p = k8s.secretToProject("default", s)
      assert.equal(p.id, "brigade-544b459e6ad7267e7791c4f77bfd1722a15e305a22cf9d3c60c5be")
      assert.equal(p.name, "github.com/deis/test-private-testbed")
      assert.equal(p.repo.name, "deis/test-private-testbed")
      assert.equal(p.repo.cloneURL, "https://github.com/deis/empty-testbed.git")
      assert.equal(p.repo.token, "pretend password\n")
      assert.equal(p.kubernetes.namespace, "default")
      assert.equal(p.kubernetes.vcsSidecar, "vcs-image:latest")
      assert.property(p.secrets, "hello")
      assert.equal(p.secrets.hello, "world")
    })
    describe("when cloneURL is missing", function() {
      it("omits cloneURL", function() {
        let s = mockSecret()
        s.data.cloneURL = ""
        let p = k8s.secretToProject("default", s)
        assert.equal(p.id, "brigade-544b459e6ad7267e7791c4f77bfd1722a15e305a22cf9d3c60c5be")
        assert.equal(p.name, "github.com/deis/test-private-testbed")
        assert.equal(p.repo.name, "deis/test-private-testbed")
        assert.equal(p.repo.token, "pretend password\n")
        assert.equal(p.kubernetes.namespace, "default")
        assert.equal(p.kubernetes.vcsSidecar, "vcs-image:latest")
        assert.property(p.secrets, "hello")
        assert.equal(p.secrets.hello, "world")

        assert.isNull(p.repo.cloneURL)
      })
    })
  })

  describe("JobRunner", function (){
    describe("when constructed", function() {
      let j: Job
      let p: Project
      let e: BrigadeEvent
      beforeEach(function(){
        j = new mock.MockJob("pequod", "whaler", ["echo hello"])
        p = mock.mockProject()
        e = mock.mockEvent()
      })
      it("creates Kubernetes objects from a job, event, and project", function() {
        let jr = new k8s.JobRunner(j, e, p)

        assert.match(jr.name, /^pequod-[0-9a-f]+-c0ffee$/)
        assert.equal(jr.runner.metadata.name, jr.name)
        assert.equal(jr.secret.metadata.name, jr.name)
        assert.equal(jr.runner.spec.containers[0].image, "whaler")

        assert.equal(jr.runner.metadata.labels.commit, e.commit)
        assert.equal(jr.secret.metadata.labels.commit, e.commit)

        assert.equal(jr.runner.metadata.labels.worker, e.workerID)
        assert.equal(jr.secret.metadata.labels.worker, e.workerID)

        assert.equal(jr.runner.metadata.labels.build, e.buildID)
        assert.equal(jr.secret.metadata.labels.build, e.buildID)

        assert.isNotNull(jr.runner.spec.containers[0].command)
        assert.property(jr.secret.data, "main.sh")
      })
      context("when env vars are specified", function() {
        beforeEach(function() {
          j.env = {"one": "first", "two": "second"}
        })
        it("sets them on the pod", function() {
          let jr = new k8s.JobRunner(j, e, p)
          let found = 0

          for (let k in j.env) {
            assert.equal(jr.secret.data[k], k8s.b64enc(j.env[k]))
            for (let env of jr.runner.spec.containers[0].env) {
              if (env.name == k) {
                assert.equal(env.valueFrom.secretKeyRef.key, k)
                found++
              }
            }
          }

          assert.equal(found, 2)
        })
      })
      context("when event is missing commit", function() {
        beforeEach(function() {
          e.commit = null
        })
        it("sets 'master' as the commit", function() {
          let jr= new k8s.JobRunner(j, e, p)

          assert.equal(jr.runner.metadata.labels.commit, "master")
          assert.equal(jr.secret.metadata.labels.commit, "master")
        })
      })
      context("when no tasks are supplied", function() {
        beforeEach(function() {
          j.tasks = []
        })
        it("omits commands", function() {
          let jr = new k8s.JobRunner(j, e, p)
          assert.isNull(jr.runner.spec.containers[0].command)
          assert.notProperty(jr.secret.data, "main.sh")
        })
      })
      context("when useSource is set to false", function() {
        beforeEach(function() {
          j.tasks = []
        })
        it("omits init container", function() {
          j.useSource = false
          let jr = new k8s.JobRunner(j, e, p)
          // Currently, annotations are only created if the init container
          // is specified.
          assert.notProperty(jr.runner.metadata, "annotations")
        })
      })
      context("when no cloneURL is set", function() {
        beforeEach(function() {
          j.tasks = []
        })
        it("omits init container", function() {
          p.repo.cloneURL = null
          let jr = new k8s.JobRunner(j, e, p)
          // Currently, annotations are only created if the init container
          // is specified.
          assert.notProperty(jr.runner.metadata, "annotations")
        })
      })
      context("when SSH key is provided", function() {
        beforeEach(function() {
          p.repo.sshKey = "SUPER SECRET"
        })
        it("attaches key to pod", function() {
          let jr = new k8s.JobRunner(j, e, p)
          let found = false
          assert.equal(jr.secret.data.brigadeSSHKey, k8s.b64enc("SUPER SECRET"))
          for (let env of jr.runner.spec.containers[0].env) {
            if (env.name == "BRIGADE_REPO_KEY") {
              found = true
              assert.equal(env.valueFrom.secretKeyRef.key, "brigadeSSHKey")
            }
          }
          assert.isTrue(found)

          // We also have to check this for the sidecar pod. But right now this is
          // embedded in a string.
          let sidecar = jr.runner.spec.initContainers[0]
          assert.equal(sidecar.env.length, 5)
          assert.equal(sidecar.env[4].name,"BRIGADE_REPO_KEY", "Has BRIGADE REPO KEY as param")
        })
      })
      context("when mount path is supplied", function() {
        beforeEach(function() {
          j.mountPath = "/ahab"
        })
        it("mounts the provided path", function() {
          let jr = new k8s.JobRunner(j, e, p)
          for (let v of jr.runner.spec.containers[0].volumeMounts) {
            if (v.name == "vcs-sidecar") {
              assert.equal(v.mountPath, j.mountPath)
            }
          }
        })
      })
      context("when cache is enabled", function() {
        beforeEach(function() {
          j.cache.enabled = true
          j.storage.enabled = true
        })
        it("configures volumes", function() {
          let jr = new k8s.JobRunner(j, e, p)
          let cname = `${ p.name.replace(/[.\/]/g, "-")}-${ j.name }`
          let foundCache = false
          let storageName = "build-storage"
          let foundStorage = false
          for (let v of jr.runner.spec.containers[0].volumeMounts) {
            if (v.name == cname) {
              foundCache = true
              assert.equal(v.mountPath, brigadeCachePath)
            } else if (v.name == storageName) {
              foundStorage = true
              assert.equal(v.mountPath, brigadeStoragePath)
            }
          }
          assert.isTrue(foundCache, "expected cache volume mount found")
          assert.isTrue(foundStorage, "expected storage volume mount found")
          foundCache = false
          foundStorage = false
          for (let v of jr.runner.spec.volumes) {
            if (v.name == cname) {
              foundCache = true
              assert.equal(v.persistentVolumeClaim.claimName, cname)
            } else if (v.name == storageName) {
              foundStorage = true
              assert.equal(v.persistentVolumeClaim.claimName, e.buildID)
            }
          }
          assert.isTrue(foundCache, "expected cache volume claim found")
          assert.isTrue(foundStorage, "expected storage volume claim found")
        })
      })
      context("when the project has enabled host mounts", function() {
        beforeEach(function() {
          p.allowHostMounts = true
        })
        it("allows jobs to mount the host's docker socket", function() {
          j.docker.enabled = true
          let jr = new k8s.JobRunner(j, e, p)
          for (let c of jr.runner.spec.containers) {
            assert.equal(c.volumeMounts.length, 3)
            var volMount = c.volumeMounts[c.volumeMounts.length-1]
            assert.equal(volMount.name, 'docker-socket')
            assert.equal(volMount.mountPath, '/var/run/docker.sock')
          }
          assert.equal(jr.runner.spec.volumes.length, 3)
          var vol = jr.runner.spec.volumes[jr.runner.spec.volumes.length-1]
          assert.equal(vol.name, 'docker-socket')
          assert.equal(vol.hostPath.path, '/var/run/docker.sock')
        })
      })
      context("when the project has disabled host mounts", function() {
        beforeEach(function() {
          p.allowHostMounts = false
        })
        it("does not allow jobs to mount the host's docker socket", function() {
          j.docker.enabled = true
          let jr = new k8s.JobRunner(j, e, p)
          for (let c of jr.runner.spec.containers) {
            assert.equal(c.volumeMounts.length, 2)
          }
          assert.equal(jr.runner.spec.volumes.length, 2)
        })
      })
      context("when job is privileged", function() {
        it("privileges containers", function() {
          j.privileged = true
          let jr = new k8s.JobRunner(j, e, p)
          for (let c of jr.runner.spec.containers) {
            assert.isTrue(c.securityContext.privileged)
          }
        })
      })
      context("when the project has privileged mode disabled", function(){
        beforeEach(function() {
          p.allowPrivilegedJobs = false
        })
        it("does not allow privileged jobs", function(){
          j.privileged = true
          let jr = new k8s.JobRunner(j, e, p)
          for (let c of jr.runner.spec.containers) {
            assert.notExists(c.securityContext.privileged)
          }
        })
      })
      context("when image pull secrets are supplied", function() {
        it("sets imagePullSecrets", function() {
          j.imagePullSecrets = ["one", "two"]
          let jr = new k8s.JobRunner(j, e, p)
          assert.equal(jr.runner.spec.imagePullSecrets.length, 2)
          for (let i = 0; i < jr.runner.spec.imagePullSecrets.length; i++) {
            let secret = jr.runner.spec.imagePullSecrets[i]
            assert.equal(secret.name, j.imagePullSecrets[i])
          }
        })
      })
      context("when a host os is supplied", function() {
        it("sets a node selector", function() {
          j.host.os = "windows"
          let jr = new k8s.JobRunner(j, e, p)
          assert.equal("windows", jr.runner.spec.nodeSelector["beta.kubernetes.io/os"])
        })
      })
      context("when a host name is supplied", function() {
        it("sets a node name", function() {
          j.host.name = "aciBridge"
          let jr = new k8s.JobRunner(j, e, p)
          assert.equal("aciBridge", jr.runner.spec.nodeName)
        })
      })
    })
  })
})

function mockSecret(): kubernetes.V1Secret {
  let s = new kubernetes.V1Secret()
  s.metadata = new kubernetes.V1ObjectMeta()
  s.data = {
    "cloneURL": "aHR0cHM6Ly9naXRodWIuY29tL2RlaXMvZW1wdHktdGVzdGJlZC5naXQ=",
    "github.token": "cHJldGVuZCBwYXNzd29yZAo=",
    "repository": "Z2l0aHViLmNvbS9kZWlzL3Rlc3QtcHJpdmF0ZS10ZXN0YmVk",
    "secrets": "eyJoZWxsbyI6ICJ3b3JsZCJ9Cg==",
    "vcsSidecar": "dmNzLWltYWdlOmxhdGVzdA=="
  }
  s.metadata.annotations = {
    "projectName": "deis/test-private-testbed"
  }

  s.metadata.labels = {
    "managedBy": "brigade",
    "release": "deis-test-private-testbed"
  }
  s.metadata.name = "brigade-544b459e6ad7267e7791c4f77bfd1722a15e305a22cf9d3c60c5be"

  return s
}
