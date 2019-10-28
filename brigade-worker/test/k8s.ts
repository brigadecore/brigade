import "mocha";
import { assert, expect } from "chai";
import * as mock from "./mock";

import * as k8s from "../src/k8s";
import { BrigadeEvent, Project } from "@brigadecore/brigadier/out/events";
import { Job, Result, brigadeCachePath, brigadeStoragePath, JobResourceLimit } from "@brigadecore/brigadier/out/job";

import * as kubernetes from "@kubernetes/client-node";

describe("k8s", function () {
  describe("b64enc", () => {
    it('encodes the string "hello"', function () {
      assert.equal(k8s.b64enc("hello"), "aGVsbG8=");
    });
  });

  describe("b64dec", () => {
    it('decodes the string "aGVsbG8="', function () {
      assert.equal(k8s.b64dec("aGVsbG8="), "hello");
    });
  });

  describe("secretToProjectVCS", function () {
    it("converts secret to project - with a VCS", function () {
      let s = mockSecretVCS();
      let p = k8s.secretToProject("default", s);
      assert.equal(
        p.id,
        "brigade-7e3d1157331f6726338395e320cffa41d2bc9e20157fd7a4df355d"
      );
      assert.equal(p.name, "github.com/brigadecore/test-private-testbed");
      assert.equal(p.repo.name, "brigadecore/test-private-testbed");
      assert.equal(
        p.repo.cloneURL,
        "https://github.com/brigadecore/empty-testbed.git"
      );
      assert.isTrue(p.repo.initGitSubmodules);
      assert.equal(p.repo.token, "pretend password\n");
      assert.equal(p.kubernetes.namespace, "default");
      assert.equal(p.kubernetes.vcsSidecar, "vcs-image:latest");
      assert.property(p.secrets, "hello");
      assert.equal(p.secrets.hello, "world");
      assert.equal(p.kubernetes.cacheStorageClass, "tashtego");
      assert.equal(p.kubernetes.buildStorageClass, "tashtego");
    });
    describe("when cloneURL is missing", function () {
      it("omits cloneURL", function () {
        let s = mockSecretVCS();
        s.data.cloneURL = "";
        let p = k8s.secretToProject("default", s);
        assert.equal(
          p.id,
          "brigade-7e3d1157331f6726338395e320cffa41d2bc9e20157fd7a4df355d"
        );
        assert.equal(p.name, "github.com/brigadecore/test-private-testbed");
        assert.equal(p.repo.name, "brigadecore/test-private-testbed");
        assert.equal(p.repo.token, "pretend password\n");
        assert.equal(p.kubernetes.namespace, "default");
        assert.equal(p.kubernetes.vcsSidecar, "vcs-image:latest");
        assert.property(p.secrets, "hello");
        assert.equal(p.secrets.hello, "world");

        assert.isNull(p.repo.cloneURL);
      });
    });
  });

  describe("secretToProjectnoVCS", function () {
    it("converts secret to project - without a VCS", function () {
      let s = mockSecretnoVCS();
      let p = k8s.secretToProject("default", s);
      assert.equal(
        p.id,
        "brigade-0e0ae80bb1243d95ea707257baebda6ccf96094cd0353d875ae903"
      );
      assert.equal(p.name, "noVCSProject");
      assert.equal(p.repo, undefined);
      assert.equal(s.data["genericGatewaySecret"], "SThkQ1g=") // Project class does not contain genericGatewaySecret field - this is because we do not want to expose the genericGatewaySecret to any job
      assert.equal(p.kubernetes.namespace, "default");
      assert.equal(p.kubernetes.vcsSidecar, "");
      assert.property(p.secrets, "hello");
      assert.equal(p.secrets.hello, "world");
      assert.equal(p.kubernetes.cacheStorageClass, "tashtego");
      assert.equal(p.kubernetes.buildStorageClass, "tashtego");
    });
  });

  describe("JobRunner", function () {
    describe("when constructed", function () {
      let j: Job;
      let p: Project;
      let e: BrigadeEvent;
      beforeEach(function () {
        j = new mock.MockJob("pequod", "whaler", ["echo hello"]);
        p = mock.mockProject();
        e = mock.mockEvent();
      });
      it("creates Kubernetes objects from a job, event, and project", function () {
        let jr = new k8s.JobRunner().init(j, e, p);

        assert.equal(jr.name, `pequod-${e.buildID}`);
        assert.equal(jr.runner.metadata.name, jr.name);
        assert.equal(jr.secret.metadata.name, jr.name);
        assert.equal(jr.runner.spec.containers[0].image, "whaler");

        assert.equal(jr.runner.metadata.labels.worker, e.workerID);
        assert.equal(jr.secret.metadata.labels.worker, e.workerID);

        assert.equal(jr.runner.metadata.labels.build, e.buildID);
        assert.equal(jr.secret.metadata.labels.build, e.buildID);

        assert.isNotNull(jr.runner.spec.containers[0].command);
        assert.property(jr.secret.data, "main.sh");
      });
      context("when env vars are specified", function () {
        context("as data", function () {
          beforeEach(function () {
            j.env = { one: "first", two: "second" };
          });
          it("sets them on the pod", function () {
            let jr = new k8s.JobRunner().init(j, e, p);
            let found = 0;

            for (let k in j.env) {
              assert.equal(jr.secret.data[k], k8s.b64enc(j.env[k] as string));
              for (let env of jr.runner.spec.containers[0].env) {
                if (env.name == k) {
                  assert.equal(env.valueFrom.secretKeyRef.key, k);
                  found++;
                }
              }
            }

            assert.equal(found, 2);
          });
        });
        context("as references", function () {
          beforeEach(function () {
            j.env = {
              one: {
                secretKeyRef: {
                  name: "secret-name",
                  key: "secret-key"
                }
              } as kubernetes.V1EnvVarSource,
              two: {
                configMapKeyRef: {
                  name: "configmap-name",
                  key: "configmap-key"
                }
              } as kubernetes.V1EnvVarSource
            };
          });
          it("sets them on the pod", function () {
            let jr = new k8s.JobRunner().init(j, e, p);
            let found = 0;

            for (let k in j.env) {
              for (let env of jr.runner.spec.containers[0].env) {
                if (env.name == k) {
                  assert.equal(env.valueFrom, j.env[k]);
                  found++;
                }
              }
            }
            assert.equal(found, 2);
          });
        });
        context("as references with allowSecretKeyRef false", function () {
          beforeEach(function () {
            j.env = {
              one: {
                secretKeyRef: {
                  name: "secret-name",
                  key: "secret-key"
                }
              } as kubernetes.V1EnvVarSource,
              two: {
                configMapKeyRef: {
                  name: "configmap-name",
                  key: "configmap-key"
                }
              } as kubernetes.V1EnvVarSource
            };
          });
          it("sets them on the pod", function () {
            let jr = new k8s.JobRunner().init(j, e, p, false);
            let found = 0;

            for (let k in j.env) {
              for (let env of jr.runner.spec.containers[0].env) {
                if (env.name == k) {
                  assert.equal(env.valueFrom, j.env[k]);
                  found++;
                }
              }
            }
            assert.equal(found, 1);
          });
        });
      });
      context("when resources are specified", function () {
        beforeEach(function () {
          j.resourceRequests.cpu = "250m";
          j.resourceRequests.memory = "512Mi";
          j.resourceLimits.cpu = "500m";
          j.resourceLimits.memory = "1Gi";
        });
        it("sets resource requests and limits for the container pod", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          let expResources = new kubernetes.V1ResourceRequirements();
          expResources.requests = { cpu: "250m", memory: "512Mi" };
          expResources.limits = { cpu: "500m", memory: "1Gi" };
          assert.deepEqual(
            jr.runner.spec.containers[0].resources,
            expResources
          );
        });
      });
      context("when service account is specified", function () {
        beforeEach(function () {
          j.serviceAccount = "svcAccount";
        });
        it("sets a service account name for the pod", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.equal(jr.runner.spec.serviceAccountName, "svcAccount");
        });
      });
      context("when no service account is specified", function () {
        it("sets a service account name to 'brigade-worker'", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.equal(jr.runner.spec.serviceAccountName, "brigade-worker");
        });
      });
      context("when custom service account is specified", function () {
        it("sets a service account name to 'custom-worker'", function () {
          k8s.options.serviceAccount = "custom-worker";
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.equal(jr.runner.spec.serviceAccountName, "custom-worker");
        });
      });
      context("when args are supplied", function () {
        beforeEach(function () {
          j.tasks = [];
          j.args = ["--aye", "-j", "kay"];
        });
        it("adds container args", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.equal(jr.runner.spec.containers[0].args.length, 3);
          assert.notProperty(jr.secret.data, "main.sh");
        });
      });
      context("when no args are supplied", function () {
        beforeEach(function () {
          j.args = [];
        });
        it("has no container args", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.notProperty(jr.runner.spec.containers[0], "args");
        });
      });
      context("when no tasks are supplied", function () {
        beforeEach(function () {
          j.tasks = [];
        });
        it("omits commands", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.isNull(jr.runner.spec.containers[0].command);
          assert.notProperty(jr.secret.data, "main.sh");
        });
      });
      context("when tasks are supplied", function () {
        beforeEach(function () {
          j.tasks = ["echo 'foo'"];
        });
        it("the resulting shell script should have certain options set", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.deepEqual(jr.runner.spec.containers[0].command, [ '/bin/sh', '/hook/main.sh' ]);
          assert.equal(k8s.b64dec(jr.secret.data["main.sh"]),
            "#!/bin/sh\n\nset -e\n\necho 'foo'");
        });
        it("the resulting bash script should have certain options set", function () {
          j.shell = "/bin/bash";
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.deepEqual(jr.runner.spec.containers[0].command, [ '/bin/bash', '/hook/main.sh' ]);
          assert.equal(k8s.b64dec(jr.secret.data["main.sh"]),
            "#!/bin/bash\n\nset -eo pipefail\n\necho 'foo'");
        });
        it("the resulting dash script shouldn't have any options set", function () {
          j.shell = "/bin/dash";
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.deepEqual(jr.runner.spec.containers[0].command, [ '/bin/dash', '/hook/main.sh' ]);
          assert.equal(k8s.b64dec(jr.secret.data["main.sh"]),
            "#!/bin/dash\n\necho 'foo'");
        });
      });
      context("when useSource is set to false", function () {
        beforeEach(function () {
          j.tasks = [];
        });
        it("omits init container", function () {
          j.useSource = false;
          let jr = new k8s.JobRunner().init(j, e, p);
          // Currently, annotations are only created if the init container
          // is specified.
          assert.deepEqual(jr.runner.metadata.annotations, {});
        });
      });
      context("when no cloneURL is set", function () {
        beforeEach(function () {
          j.tasks = [];
        });
        it("omits init container", function () {
          p.repo.cloneURL = null;
          let jr = new k8s.JobRunner().init(j, e, p);
          // Currently, annotations are only created if the init container
          // is specified.
          assert.deepEqual(jr.runner.metadata.annotations, {});
        });
      });
      context("when SSH key is provided", function () {
        beforeEach(function () {
          p.repo.sshKey = "SUPER SECRET";
        });
        it("attaches key to pod", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          let sidecar = jr.runner.spec.initContainers[0];
          assert.equal(sidecar.env.length, 14);

          let hasBrigadeRepoKey: boolean = false;
          for (let i of sidecar.env) {
            if (i.name === "BRIGADE_REPO_KEY") {
              hasBrigadeRepoKey = true;
              break;
            }
          }
          assert.isTrue(hasBrigadeRepoKey, "Has BRIGADE REPO KEY as param");
        });
      });
      context("when sidecar is disabled", function () {
        beforeEach(function () {
          p.kubernetes.vcsSidecar = "";
        });
        it("job runner should have no init containers", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.equal(jr.runner.spec.initContainers.length, 0);
        });
        it("job runner should have no sidecar volumes", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.notDeepInclude(
            jr.runner.spec.volumes,
            { name: "vcs-sidecar", emptyDir: {} } as kubernetes.V1Volume
          );
        });
        it("job runner should have no sidecar volume mounts", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.notDeepInclude(
            jr.runner.spec.containers[0].volumeMounts,
            { name: "vcs-sidecar", mountPath: j.mountPath } as kubernetes.V1VolumeMount
          );
        });
      });
      context("when mount path is supplied", function () {
        beforeEach(function () {
          j.mountPath = "/ahab";
        });
        it("mounts the provided path", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          for (let v of jr.runner.spec.containers[0].volumeMounts) {
            if (v.name == "vcs-sidecar") {
              assert.equal(v.mountPath, j.mountPath);
            }
          }
        });
      });
      context("when cache is enabled", function () {
        beforeEach(function () {
          j.cache.enabled = true;
          j.storage.enabled = true;
        });
        it("configures volumes", function () {
          // We uppercase to test that names are correctly downcased. Issue #224
          j.name = j.name.toUpperCase();
          let jr = new k8s.JobRunner().init(j, e, p);
          let cname = `${p.name.replace(
            /[.\/]/g,
            "-"
          )}-${j.name.toLowerCase()}`;
          let foundCache = false;
          let storageName = "build-storage";
          let foundStorage = false;
          for (let v of jr.runner.spec.containers[0].volumeMounts) {
            if (v.name == cname) {
              foundCache = true;
              assert.equal(v.mountPath, brigadeCachePath);
            } else if (v.name == storageName) {
              foundStorage = true;
              assert.equal(v.mountPath, brigadeStoragePath);
            }
          }
          assert.isTrue(foundCache, "expected cache volume mount found");
          assert.isTrue(foundStorage, "expected storage volume mount found");
          foundCache = false;
          foundStorage = false;
          for (let v of jr.runner.spec.volumes) {
            if (v.name == cname) {
              foundCache = true;
              assert.equal(v.persistentVolumeClaim.claimName, cname);
            } else if (v.name == storageName) {
              foundStorage = true;
              assert.equal(
                v.persistentVolumeClaim.claimName,
                e.workerID.toLowerCase()
              );
            }
          }
          assert.isTrue(foundCache, "expected cache volume claim found");
          assert.isTrue(foundStorage, "expected storage volume claim found");
        });
        it("configures volumes with custom paths", function () {
          j.cache.path = "/cache";
          j.cache.enabled = true;
          j.storage.path = "/storage";
          j.storage.enabled = true;
          let jr = new k8s.JobRunner().init(j, e, p);

          let cname = `${p.name.replace(
            /[.\/]/g,
            "-"
          )}-${j.name.toLowerCase()}`;
          let foundCache = false;
          let storageName = "build-storage";
          let foundStorage = false;
          for (let v of jr.runner.spec.containers[0].volumeMounts) {
            if (v.name == cname) {
              foundCache = true;
              assert.equal(v.mountPath, "/cache");
            } else if (v.name == storageName) {
              foundStorage = true;
              assert.equal(v.mountPath, "/storage");
            }
          }
          assert.isTrue(foundCache, "expected cache volume mount found");
          assert.isTrue(foundStorage, "expected storage volume mount found");
        });
      });
      context("when the project has enabled host mounts", function () {
        beforeEach(function () {
          p.allowHostMounts = true;
        });
        it("allows jobs to mount the host's docker socket", function () {
          j.docker.enabled = true;
          let jr = new k8s.JobRunner().init(j, e, p);
          for (let c of jr.runner.spec.containers) {
            assert.equal(c.volumeMounts.length, 3);
            var volMount = c.volumeMounts[c.volumeMounts.length - 1];
            assert.equal(volMount.name, "docker-socket");
            assert.equal(volMount.mountPath, "/var/run/docker.sock");
          }
          assert.equal(jr.runner.spec.volumes.length, 3);
          var vol = jr.runner.spec.volumes[jr.runner.spec.volumes.length - 1];
          assert.equal(vol.name, "docker-socket");
          assert.equal(vol.hostPath.path, "/var/run/docker.sock");
        });
      });
      context("when the project has disabled host mounts", function () {
        beforeEach(function () {
          p.allowHostMounts = false;
        });
        it("does not allow jobs to mount the host's docker socket", function () {
          j.docker.enabled = true;
          let jr = new k8s.JobRunner().init(j, e, p);
          for (let c of jr.runner.spec.containers) {
            assert.equal(c.volumeMounts.length, 2);
          }
          assert.equal(jr.runner.spec.volumes.length, 2);
        });
      });
      context("when a hostPath type volume is set for a job", function () {
        beforeEach(function () {
          var v = new kubernetes.V1Volume();
          v.name = "mock-volume";
          v.hostPath = {
            path: "/some/path",
            type: "Directory"
          };
          j.volumes.push(v);
        });
        it("with allowHostMounts disabled, error is thrown", function () {
          expect(() => new k8s.JobRunner().init(j, e, p)).to.throw(Error, "allowHostMounts is false in this project, not mounting /some/path");
        });
        it("with allowHostMounts enabled, no error is thrown", function () {
          p.allowHostMounts = true;
          expect(() => new k8s.JobRunner().init(j, e, p)).to.not.throw(Error);
        });
        it("all properties are properly set", function () {
          p.allowHostMounts = true;
          expect(() => new k8s.JobRunner().init(j, e, p)).to.not.throw(Error);
          let jr = new k8s.JobRunner().init(j, e, p);

          assert.equal(jr.runner.spec.volumes[2].name, "mock-volume");
          assert.equal(jr.runner.spec.volumes[2].hostPath.path, "/some/path");
          assert.equal(jr.runner.spec.volumes[2].hostPath.type, "Directory");
        });

        it("and job enables Docker, all properties are properly set", function () {
          p.allowHostMounts = true;
          j.docker.enabled = true;
          expect(() => new k8s.JobRunner().init(j, e, p)).to.not.throw(Error);
          let jr = new k8s.JobRunner().init(j, e, p);

          assert.equal(jr.runner.spec.volumes[2].name, "docker-socket");
          assert.equal(jr.runner.spec.volumes[2].hostPath.path, "/var/run/docker.sock");

          assert.equal(jr.runner.spec.volumes[3].name, "mock-volume");
          assert.equal(jr.runner.spec.volumes[3].hostPath.path, "/some/path");
          assert.equal(jr.runner.spec.volumes[3].hostPath.type, "Directory");
        });
      });

      context("when a persistent volume type volume is set for a job", function () {
        beforeEach(function () {
          var v = new kubernetes.V1Volume();
          v.name = "mock-volume";
          v.persistentVolumeClaim = {
            claimName: "some-claim"
          };
          j.volumes.push(v);
        });
        it("with allowHostMounts disabled, no error is thrown", function () {
          expect(() => new k8s.JobRunner().init(j, e, p)).to.not.throw(Error);
        });

        it("with allowHostMounts enabled, no error is thrown", function () {
          p.allowHostMounts = true;
          expect(() => new k8s.JobRunner().init(j, e, p)).to.not.throw(Error);
        });
        it("all properties are properly set", function () {
          expect(() => new k8s.JobRunner().init(j, e, p)).to.not.throw(Error);
          let jr = new k8s.JobRunner().init(j, e, p);

          assert.equal(jr.runner.spec.volumes[2].name, "mock-volume");
          assert.equal(jr.runner.spec.volumes[2].persistentVolumeClaim.claimName, "some-claim");
        });
        it("and job enables Docker, all properties are properly set", function () {
          p.allowHostMounts = true;
          j.docker.enabled = true;
          expect(() => new k8s.JobRunner().init(j, e, p)).to.not.throw(Error);
          let jr = new k8s.JobRunner().init(j, e, p);

          assert.equal(jr.runner.spec.volumes[2].name, "docker-socket");
          assert.equal(jr.runner.spec.volumes[2].hostPath.path, "/var/run/docker.sock");

          assert.equal(jr.runner.spec.volumes[3].name, "mock-volume");
          assert.equal(jr.runner.spec.volumes[3].persistentVolumeClaim.claimName, "some-claim");
        });
      });

      context("when volumeMounts is set for a job", function () {
        beforeEach(function () {
          var v = new kubernetes.V1Volume();
          v.name = "mock-volume";
          v.persistentVolumeClaim = {
            claimName: "some-claim"
          };
          j.volumes.push(v);

          var m = new kubernetes.V1VolumeMount();
          m.name = "mock-volume"
          m.mountPath = "/mock/volume";
          j.volumeMounts.push(m);
        });
        it("the volume mounts are set in all containers", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          for (let c of jr.runner.spec.containers) {
            assert.equal(c.volumeMounts[2].mountPath, "/mock/volume");
            assert.equal(c.volumeMounts[2].name, "mock-volume");
          }
        });
        it("and Docker is enabled for the job, the volume mounts are set in all containers", function () {
          p.allowHostMounts = true;
          j.docker.enabled = true;
          let jr = new k8s.JobRunner().init(j, e, p);
          for (let c of jr.runner.spec.containers) {
            assert.equal(c.volumeMounts[2].mountPath, "/var/run/docker.sock");
            assert.equal(c.volumeMounts[2].name, "docker-socket");

            assert.equal(c.volumeMounts[3].mountPath, "/mock/volume");
            assert.equal(c.volumeMounts[3].name, "mock-volume");
          }
        });
      });
      context("when a volumeMount is set for a job, but the referenced volume does not exist", function () {
        it("error is thrown", function () {
          var m = new kubernetes.V1VolumeMount();
          m.name = "mock-volume"
          m.mountPath = "/mock/volume";
          j.volumeMounts.push(m);

          expect(() => new k8s.JobRunner().init(j, e, p)).to.throw(Error, "volume mock-volume referenced in volume mount is not defined");
        });
      });
      context("when job is privileged", function () {
        it("privileges containers", function () {
          j.privileged = true;
          let jr = new k8s.JobRunner().init(j, e, p);
          for (let c of jr.runner.spec.containers) {
            assert.isTrue(c.securityContext.privileged);
          }
        });
      });
      context("when the project has privileged mode disabled", function () {
        beforeEach(function () {
          p.allowPrivilegedJobs = false;
        });
        it("does not allow privileged jobs", function () {
          j.privileged = true;
          let jr = new k8s.JobRunner().init(j, e, p);
          for (let c of jr.runner.spec.containers) {
            assert.notExists(c.securityContext.privileged);
          }
        });
      });
      context("when image pull secrets are supplied", function () {
        it("sets imagePullSecrets", function () {
          j.imagePullSecrets = ["one", "two"];
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.equal(jr.runner.spec.imagePullSecrets.length, 2);
          for (let i = 0; i < jr.runner.spec.imagePullSecrets.length; i++) {
            let secret = jr.runner.spec.imagePullSecrets[i];
            assert.equal(secret.name, j.imagePullSecrets[i]);
          }
        });
      });
      context("when a host os is supplied", function () {
        it("sets a node selector", function () {
          j.host.os = "windows";
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.equal(
            "windows",
            jr.runner.spec.nodeSelector["beta.kubernetes.io/os"]
          );
        });
      });
      context("when a host name is supplied", function () {
        it("sets a node name", function () {
          j.host.name = "aciBridge";
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.equal("aciBridge", jr.runner.spec.nodeName);
        });
      });
      context("when host nodeSelector are supplied", function () {
        it("sets a node selector", function () {
          j.host.nodeSelector.set("inn", "spouter");
          j.host.nodeSelector.set("ship", "pequod");
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.equal("spouter", jr.runner.spec.nodeSelector["inn"]);
          assert.equal("pequod", jr.runner.spec.nodeSelector["ship"]);
        });
      });
      context("when vcsSidecar resources defined", function () {
        beforeEach(function () {
          p.kubernetes.vcsSidecarResourcesLimitsCPU = "100m";
          p.kubernetes.vcsSidecarResourcesLimitsMemory = "100Mi";
          p.kubernetes.vcsSidecarResourcesRequestsCPU = "50m";
          p.kubernetes.vcsSidecarResourcesRequestsMemory = "50Mi";
        });
        it("sets resource requests and limits for the init-container pod", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          let expResources = new kubernetes.V1ResourceRequirements();
          expResources.limits = { cpu: "100m", memory: "100Mi" };
          expResources.requests = { cpu: "50m", memory: "50Mi" };
          assert.deepEqual(
            jr.runner.spec.initContainers[0].resources,
            expResources
          );
        });
      });
      context("when vcsSidecar only cpu resources defined", function () {
        beforeEach(function () {
          p.kubernetes.vcsSidecarResourcesLimitsCPU = "100m";
          p.kubernetes.vcsSidecarResourcesRequestsCPU = "50m";
        });
        it("sets only cpu resource requests and limits for the init-container pod", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          let expResources = new kubernetes.V1ResourceRequirements();
          expResources.limits = { cpu: "100m" };
          expResources.requests = { cpu: "50m" };
          assert.deepEqual(
            jr.runner.spec.initContainers[0].resources,
            expResources
          );
        });
      });
      context("when vcsSidecar only memory resources defined", function () {
        beforeEach(function () {
          p.kubernetes.vcsSidecarResourcesLimitsMemory = "100Mi";
          p.kubernetes.vcsSidecarResourcesRequestsMemory = "50Mi";
        });
        it("sets only memory resource requests and limits for the init-container pod", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          let expResources = new kubernetes.V1ResourceRequirements();
          expResources.limits = { memory: "100Mi" };
          expResources.requests = { memory: "50Mi" };
          assert.deepEqual(
            jr.runner.spec.initContainers[0].resources,
            expResources
          );
        });
      });
      context("when no job shell is specified", function () {
        it("default shell is /bin/sh", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.deepEqual(jr.runner.spec.containers[0].command, ['/bin/sh', '/hook/main.sh']);
        });
      });
      context("when job shell is specified", function () {
        beforeEach(function () {
          j.shell = "/bin/bash"
        });
        it("shell is /bin/bash", function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          assert.deepEqual(jr.runner.spec.containers[0].command, ['/bin/bash', '/hook/main.sh']);
        });
      });
      context("when logs is called", function() {
        it("when the job has been canceled", async function () {
          let jr = new k8s.JobRunner().init(j, e, p);
          jr.cancel = true;
          let logs = await jr.logs();
          assert.equal("pod pequod-1234567890abcdef still unscheduled or pending when job was canceled; no logs to return.", logs);
        });
      });
    });
    describe("cachePVC", () => {
      let jr: k8s.JobRunner;
      beforeEach(function () {
        let j = new mock.MockJob("pequod", "whaler", ["echo hello"]);
        let p = mock.mockProject();
        let e = mock.mockEvent();
        jr = new k8s.JobRunner().init(j, e, p);
      });
      context("when global default cache storage class is specified", () => {
        beforeEach(function () {
          jr.options.defaultCacheStorageClass = "foo";
        });
        context("when the cache storage class is overridden at the project level", () => {
          beforeEach(function () {
            jr.project.kubernetes.cacheStorageClass = "bar";
          });
          it("it uses that", () => {
            let pvc = jr['cachePVC']()
            assert.equal(pvc.spec.storageClassName, jr.project.kubernetes.cacheStorageClass)
          });
        });
        context("when the cache storage class is not overridden at the project level", () => {
          beforeEach(function () {
            jr.project.kubernetes.cacheStorageClass = "";
          });
          it("it falls back on the global default", () => {
            let pvc = jr['cachePVC']()
            assert.equal(pvc.spec.storageClassName, jr.options.defaultCacheStorageClass)
          });
        });
      });
      context("when global default cache storage class is not specified", () => {
        beforeEach(function () {
          jr.options.defaultCacheStorageClass = "";
        });
        context("when the cache storage class is overridden at the project level", () => {
          beforeEach(function () {
            jr.project.kubernetes.cacheStorageClass = "bar";
          });
          it("it uses that", () => {
            let pvc = jr['cachePVC']()
            assert.equal(pvc.spec.storageClassName, jr.project.kubernetes.cacheStorageClass)
          });
        });
        context("when the cache storage class is not overridden at the project level", () => {
          beforeEach(function () {
            jr.project.kubernetes.cacheStorageClass = "";
          });
          it("it falls back on the cluster default", () => {
            let pvc = jr['cachePVC']()
            // Undefined means k8s will use the cluster's default storage class
            assert.isUndefined(pvc.spec.storageClassName);
          });
        });
      });
    });
  });

  describe("BuildStorage", () => {
    describe("buildPVC", () => {
      let bs: k8s.BuildStorage;
      beforeEach(function () {
        bs = new k8s.BuildStorage();
        bs.proj = mock.mockProject();
      });
      context("when global default build storage class is specified", () => {
        beforeEach(function () {
          bs.options.defaultBuildStorageClass = "foo";
        });
        context("when the build storage class is overridden at the project level", () => {
          beforeEach(function () {
            bs.proj.kubernetes.buildStorageClass = "bar";
          });
          it("it uses that", () => {
            let pvc = bs['buildPVC']("10Gi");
            assert.equal(pvc.spec.storageClassName, bs.proj.kubernetes.buildStorageClass)
          });
        });
        context("when the build storage class is not overridden at the project level", () => {
          beforeEach(function () {
            bs.proj.kubernetes.buildStorageClass = "";
          });
          it("it falls back on the global default", () => {
            let pvc = bs['buildPVC']("10Gi");
            assert.equal(pvc.spec.storageClassName, bs.options.defaultBuildStorageClass)
          });
        });
      });
      context("when global default build storage class is not specified", () => {
        beforeEach(function () {
          bs.options.defaultBuildStorageClass = "";
        });
        context("when the build storage class is overridden at the project level", () => {
          beforeEach(function () {
            bs.proj.kubernetes.buildStorageClass = "bar";
          });
          it("it uses that", () => {
            let pvc = bs['buildPVC']("10Gi");
            assert.equal(pvc.spec.storageClassName, bs.proj.kubernetes.buildStorageClass)
          });
        });
        context("when the build storage class is not overridden at the project level", () => {
          beforeEach(function () {
            bs.proj.kubernetes.buildStorageClass = "";
          });
          it("it falls back on the cluster default", () => {
            let pvc = bs['buildPVC']("10Gi");
            // Undefined means k8s will use the cluster's default storage class
            assert.isUndefined(pvc.spec.storageClassName);
          });
        });
      });
    });
  });
});

function mockSecretVCS(): kubernetes.V1Secret {
  let s = new kubernetes.V1Secret();
  s.metadata = new kubernetes.V1ObjectMeta();
  s.data = {
    cloneURL: "aHR0cHM6Ly9naXRodWIuY29tL2JyaWdhZGVjb3JlL2VtcHR5LXRlc3RiZWQuZ2l0",
    initGitSubmodules: "dHJ1ZQ==",
    "github.token": "cHJldGVuZCBwYXNzd29yZAo=",
    repository: "Z2l0aHViLmNvbS9icmlnYWRlY29yZS90ZXN0LXByaXZhdGUtdGVzdGJlZA==",
    secrets: "eyJoZWxsbyI6ICJ3b3JsZCJ9Cg==",
    vcsSidecar: "dmNzLWltYWdlOmxhdGVzdA==",
    buildStorageSize: "NTBNaQ==",
    "kubernetes.cacheStorageClass": "dGFzaHRlZ28=",
    "kubernetes.buildStorageClass": "dGFzaHRlZ28="
  };
  s.metadata.annotations = {
    projectName: "brigadecore/test-private-testbed"
  };

  s.metadata.labels = {
    managedBy: "brigade",
    release: "brigadecore-test-private-testbed"
  };
  s.metadata.name =
    "brigade-7e3d1157331f6726338395e320cffa41d2bc9e20157fd7a4df355d";

  return s;
}

function mockSecretnoVCS(): kubernetes.V1Secret {
  let s = new kubernetes.V1Secret();
  s.metadata = new kubernetes.V1ObjectMeta();
  s.data = {
    secrets: "eyJoZWxsbyI6ICJ3b3JsZCJ9Cg==",
    buildStorageSize: "NTBNaQ==",
    "kubernetes.cacheStorageClass": "dGFzaHRlZ28=",
    "kubernetes.buildStorageClass": "dGFzaHRlZ28=",
    genericGatewaySecret: "SThkQ1g="
  };
  s.metadata.annotations = {
    projectName: "noVCSProject"
  };
  s.metadata.labels = {
    managedBy: "brigade",
    release: "brigadecore-test-private-testbed"
  };
  s.metadata.name =
    "brigade-0e0ae80bb1243d95ea707257baebda6ccf96094cd0353d875ae903";

  return s;
}

function checkObjWithVal(prop, val, arr: Array<any>): boolean {
  return arr.some(obj => obj[prop] === val);
} 