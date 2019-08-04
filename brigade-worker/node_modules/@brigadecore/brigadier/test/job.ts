import "mocha";
import { assert } from "chai";
import * as mock from "./mock";

import {
  JobHost,
  JobCache,
  JobStorage,
  brigadeCachePath,
  brigadeStoragePath,
  jobNameIsValid,
  JobVolumeConfig
} from "../src/job";
import { V1Volume, V1VolumeMount } from "@kubernetes/client-node/dist/api";

describe("job", function () {
  describe("jobNameIsValid", () => {
    it("allows DNS-like names", function () {
      let legal = ["abcdef", "ab", "a-b", "a-9", "a12345678b", "a.b"];
      for (let n of legal) {
        assert.isTrue(jobNameIsValid(n), "tested " + n);
      }
    });
    it("disallows non-DNS-like names", function () {
      let illegal = [
        "ab-", // no trailing -
        "-ab", // no leading dash
        "a_b", // underscore is illegal
        "ab.", // trailing dot is illegal
        "A-B" // Capitals are illegal
      ];
      for (let n of illegal) {
        assert.isFalse(jobNameIsValid(n), "tested " + n);
      }
    });
  });

  describe("JobCache", function () {
    describe("#constructor", function () {
      it("correctly sets default values", function () {
        let c = new JobCache();
        assert.equal(c.path, brigadeCachePath, "Dir is /brigade/cache");
        assert.isFalse(c.enabled, "disabled by default");
        assert.equal(c.size, "5Mi", "size is 5mi");
      });
    });
    describe("#setPath", function () {
      it("correctly sets and gets path", function () {
        let c = new JobCache();
        assert.equal(c.path, brigadeCachePath, "Dir is /brigade/cache");
        c.path = "/foo"
        assert.equal(c.path, "/foo", "Assert dir is /foo");
      });
    });
  });
  describe("JobStorage", function () {
    describe("#constructor", function () {
      it("correctly sets default values", function () {
        let c = new JobStorage();
        assert.equal(
          c.path,
          brigadeStoragePath,
          "Dir is " + brigadeStoragePath
        );
        assert.isFalse(c.enabled, "disabled by default");
      });
    });
    describe("#setPath", function () {
      it("correctly sets and gets path", function () {
        let s = new JobStorage();
        assert.equal(s.path, brigadeStoragePath, "Dir is /brigade/cache");
        s.path = "/foo"
        assert.equal(s.path, "/foo", "Assert dir is /foo");
      });
    });
  });
  describe("JobHost", function () {
    describe("#constructor", function () {
      it("correctly sets default values", function () {
        let h = new JobHost();
        assert.equal(0, h.nodeSelector.size, "had empty nodeSelector map");
        // Validate that the nodeSelector structure works like a map.
        h.nodeSelector.set("callMe", "Ishmael");
        assert.equal("Ishmael", h.nodeSelector.get("callMe"));
      });
    });
  });
  describe("Job", function () {
    // MockJob extends Job, but overrides the empty run() and logs() calls
    let j: mock.MockJob;
    describe("#constructor", function () {
      it("creates a named job", function () {
        j = new mock.MockJob("my-name");
        assert.equal(j.name, "my-name");
      });
      it("starts with initialized JobHost", function () {
        j = new mock.MockJob("name");
        assert.equal(j.host.nodeSelector.size, 0);
      });
      context("when image is supplied", function () {
        it("sets image property", function () {
          j = new mock.MockJob("my-name", "alpine:3.4");
          assert.equal(j.image, "alpine:3.4");
        });
      });
      context("when imageForcePull is supplied", function () {
        it("sets imageForcePull property", function () {
          j = new mock.MockJob("my-name", "alpine:3.4", [], true);
          assert.isTrue(j.imageForcePull);
        });
      });
      context("when tasks are supplied", function () {
        it("sets task list", function () {
          j = new mock.MockJob("my", "img", ["a", "b", "c"]);
          assert.deepEqual(j.tasks, ["a", "b", "c"]);
        });
      });
      context("when serviceAccount is supplied", function () {
        it("sets serviceAccount property", function () {
          j = new mock.MockJob("my-name", "alpine:3.4", [], true);
          j.serviceAccount = "svcAccount";
          assert.equal(j.serviceAccount, "svcAccount");
        });
      });
    });
    describe("#podName", function () {
      beforeEach(function () {
        j = new mock.MockJob("my-job");
      });
      context("before run", function () {
        it("is empty", function () {
          assert.isEmpty(j.podName);
        });
      });
      context("after run", function () {
        it("is accessible", function (done) {
          j.run().then(rez => {
            assert.equal(j.podName, "generated-fake-job-name");
            done();
          });
        });
      });
    });
    describe("#cache", function () {
      it("is disabled by default", function () {
        assert.isFalse(j.cache.enabled);
      });
    });
    describe("#storage", function () {
      it("is disabled by default", function () {
        assert.isFalse(j.storage.enabled);
      });
    });
    describe("#annotations", function () {
      beforeEach(function () {
        j = new mock.MockJob("my-job");
      });
      it("is an empty list that can be written", function () {
        assert.deepEqual(j.annotations, {});
        j.annotations['some_kubetoiam/thing'] = 'my/path';
        assert.deepEqual(j.annotations, { 'some_kubetoiam/thing': 'my/path' });
      });
    });
    describe("#volumeConfig", function () {
      beforeEach(function () {
        j = new mock.MockJob("my-job");
        var m = new V1VolumeMount();
        m.name = "mock-volume"
        m.mountPath = "/mock/volume";

        var v = new V1Volume();
        v.name = "mock-volume";
        v.hostPath = {
          path: "/some/path",
          type: "Directory"
        };

        j.volumeConfig = [
          new JobVolumeConfig(m, v)
        ];
      });
      it("is correctly configured", function () {
        assert.equal(j.volumeConfig.length, 1);
        assert.equal(j.volumeConfig[0].mount!.name, "mock-volume");
        assert.equal(j.volumeConfig[0].mount!.mountPath, "/mock/volume");

        assert.equal(j.volumeConfig[0].volume!.name, "mock-volume");
        assert.equal(j.volumeConfig[0].volume!.hostPath!.path, "/some/path");
        assert.equal(j.volumeConfig[0].volume!.hostPath!.type, "Directory");
      });
    });
  });
});
