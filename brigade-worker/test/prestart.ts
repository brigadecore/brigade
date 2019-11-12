import "mocha";
import { assert } from "chai";
import * as mock from "mock-require";
import * as sinon from "sinon";
import { getDefaultCompilerOptions } from "typescript";
import { write, read } from "fs";

// Using `let` and `require` so `mock.reRequire` is legal later.
let prestart = require('../prestart');

describe("prestart", function() {
  describe("buildPackageList", function() {
    it("rejects null", function () {
      assert.throws(() => prestart.buildPackageList(null));
    });

    it("builds a list with no deps", function() {
      assert.deepEqual(
        prestart.buildPackageList({}),
        []);
    });

    it("builds a list with multiple deps", function() {
      assert.deepEqual(
        prestart.buildPackageList({
          "is-thirteen": "2.0.0",
          "lodash": "4.0.0",
        }),
        ["is-thirteen@2.0.0", "lodash@4.0.0"]);
    });
  });

  describe("addYarn", function() {

    it("rejects null", function() {
      assert.throws(() => prestart.addYarn(null));
    });

    it("rejects the empty list", function() {
      assert.throws(() => prestart.addYarn([]));
    });

    describe("mocked", function() {
      let execFileSync;
      beforeEach(function() {
        execFileSync = sinon.stub();
        mock("child_process", { execFileSync });
        prestart = mock.reRequire("../prestart");
      });

      afterEach(function() {
        mock.stopAll();
      });

      it("invokes execFileSync for a single package", function () {
        try {
          prestart.addYarn(["is-thirteen@2.0.0"]);
        } finally {
          mock.stopAll();
        }

        sinon.assert.calledOnce(execFileSync);
        sinon.assert.calledWithExactly(
          execFileSync, "yarn", ["add", "is-thirteen@2.0.0"]);
      });

      it("invokes execFileSync for multiple packages", function () {
        try {
          prestart.addYarn(["is-thirteen@2.0.0", "lodash@4.0.0"]);
        } finally {
          mock.stopAll();
        }

        sinon.assert.calledOnce(execFileSync);
        sinon.assert.calledWithExactly(
          execFileSync, "yarn", ["add", "is-thirteen@2.0.0", "lodash@4.0.0"]);
      });

    });
  });

  describe("createConfig", function () {
    let
      existsSync: sinon.SinonStub,
      readFileSync: sinon.SinonStub,
      writeFileSync: sinon.SinonStub,
      exit: sinon.SinonStub;

      beforeEach(function() {
        existsSync = sinon.stub();
        readFileSync = sinon.stub();
        readFileSync.callsFake(() => "{}");
        writeFileSync = sinon.stub();

        mock("fs", { existsSync, readFileSync, writeFileSync })

        prestart = mock.reRequire("../prestart");
      });

      afterEach(function() {
        mock.stopAll();
      });

      it("no brigade.json", function() {
        existsSync.callsFake(() => false);

        prestart.createConfig();

        sinon.assert.calledTwice(existsSync);
        sinon.assert.calledWith(existsSync.firstCall, prestart.mountedConfigFile);
        sinon.assert.calledWith(existsSync.secondCall, prestart.vcsConfigFile);
        sinon.assert.notCalled(writeFileSync);
      });

      it("config exists via mounted file", function() {
        existsSync.callsFake((...args) => {
          return args[0] === prestart.mountedConfigFile;
        });

        prestart.createConfig();

        sinon.assert.calledOnce(existsSync);
        sinon.assert.calledWith(existsSync.firstCall, prestart.mountedConfigFile);
        sinon.assert.called(writeFileSync);
        sinon.assert.calledWithExactly(writeFileSync, prestart.configFile, "{}")
      });

      it("no brigade.json mounted, but exists in vcs", function() {
        existsSync.callsFake((...args) => {
          return args[0] === prestart.vcsConfigFile;
        });

        prestart.createConfig();

        sinon.assert.calledTwice(existsSync);
        sinon.assert.calledWith(existsSync.firstCall, prestart.mountedConfigFile);
        sinon.assert.calledWith(existsSync.secondCall, prestart.vcsConfigFile);
        sinon.assert.called(writeFileSync);
        sinon.assert.calledWithExactly(writeFileSync, prestart.configFile, "{}")
      });
    });

  describe("addDeps", function () {
    let
      execFileSync: sinon.SinonStub,
      existsSync: sinon.SinonStub,
      readFileSync: sinon.SinonStub,
      writeFileSync: sinon.SinonStub,
      exit: sinon.SinonStub;

      beforeEach(function() {
        execFileSync = sinon.stub();
        mock("child_process", { execFileSync });

        existsSync = sinon.stub();
        readFileSync = sinon.stub();
        writeFileSync = sinon.stub();
        mock("fs", { existsSync, readFileSync, writeFileSync })

        exit = sinon.stub();
        mock("process", { env: {}, exit });

        sinon.stub(console, 'error');

        prestart = mock.reRequire("../prestart");
      });

      afterEach(function() {
        mock.stopAll();

        (console as any).error.restore();
      });

      it("no config file exists", function() {
        existsSync.callsFake(() => false);

        prestart.addDeps();

        sinon.assert.calledThrice(existsSync);
        sinon.assert.calledWith(existsSync.firstCall, prestart.mountedConfigFile);
        sinon.assert.calledWith(existsSync.secondCall, prestart.vcsConfigFile);
        sinon.assert.calledWith(existsSync.thirdCall, prestart.configFile);
        sinon.assert.notCalled(execFileSync);
        sinon.assert.notCalled(exit);
      });

      it("no dependencies object", function() {
        mock(prestart.configFile, {});
        existsSync.callsFake(() => true);

        prestart.addDeps();

        sinon.assert.calledTwice(existsSync);
        sinon.assert.calledWith(existsSync.firstCall, prestart.mountedConfigFile);
        sinon.assert.calledWith(existsSync.secondCall, prestart.configFile);
        sinon.assert.notCalled(execFileSync);
        sinon.assert.notCalled(exit);
      });

      it("empty dependencies", function() {
        mock(prestart.configFile, { dependencies: {}})
        existsSync.callsFake(() => true);

        prestart.addDeps();

        sinon.assert.calledTwice(existsSync);
        sinon.assert.calledWith(existsSync.firstCall, prestart.mountedConfigFile);
        sinon.assert.calledWith(existsSync.secondCall, prestart.configFile);
        sinon.assert.notCalled(execFileSync);
        sinon.assert.notCalled(exit);
      });

      it("one dependency", function() {
        mock(prestart.configFile, {
          dependencies: {
            "is-thirteen": "2.0.0",
          },
        });
        existsSync.callsFake(() => true);

        prestart.addDeps();

        sinon.assert.calledTwice(existsSync);
        sinon.assert.calledWith(existsSync.firstCall, prestart.mountedConfigFile);
        sinon.assert.calledWith(existsSync.secondCall, prestart.configFile);
        sinon.assert.calledOnce(execFileSync);
        sinon.assert.calledWithExactly(
          execFileSync, "yarn", ["add", "is-thirteen@2.0.0"]);
        sinon.assert.notCalled(exit);
      })

      it("two dependencies", function() {
        mock(prestart.configFile, {
          dependencies: {
            "is-thirteen": "2.0.0",
            "lodash": "4.0.0",
          },
        });
        existsSync.callsFake(() => true);

        prestart.addDeps();

        sinon.assert.calledTwice(existsSync);
        sinon.assert.calledWith(existsSync.firstCall, prestart.mountedConfigFile);
        sinon.assert.calledWith(existsSync.secondCall, prestart.configFile);
        sinon.assert.calledOnce(execFileSync);
        sinon.assert.calledWithExactly(
          execFileSync, "yarn", ["add", "is-thirteen@2.0.0", "lodash@4.0.0"]);
        sinon.assert.notCalled(exit);
      });

      it("yarn error", function() {
        mock(prestart.configFile, {
          dependencies: {
            "is-thirteen": "2.0.0",
          },
        });
        existsSync.callsFake(() => true);
        execFileSync.callsFake(() => {
          const e = new Error('Command failed: yarn');
          (e as any).status = 1;
          throw e;
        });

        prestart.addDeps();

        sinon.assert.calledTwice(existsSync);
        sinon.assert.calledWith(existsSync.firstCall, prestart.mountedConfigFile);
        sinon.assert.calledWith(existsSync.secondCall, prestart.configFile);
        sinon.assert.calledOnce(execFileSync);
        sinon.assert.calledWithExactly(execFileSync, "yarn", ["add", "is-thirteen@2.0.0"]);
        sinon.assert.calledOnce(exit);
        sinon.assert.calledWithExactly(exit, 1);
      });
  });
});

