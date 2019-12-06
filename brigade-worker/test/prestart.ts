import "mocha";
import { assert } from "chai";
import * as mock from "mock-require";
import * as sinon from "sinon";

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
      writeFileSync: sinon.SinonStub;

      beforeEach(function() {
        existsSync = sinon.stub();
        readFileSync = sinon.stub();
        readFileSync.callsFake(() => "{}");
        writeFileSync = sinon.stub();

        mock("fs", { existsSync, readFileSync, writeFileSync });

        prestart = mock.reRequire("../prestart");
      });

      afterEach(function() {
        mock.stopAll();
      });

      it("no brigade.json", function() {
        existsSync.callsFake(() => false);

        prestart.createConfig();

        assert.equal(existsSync.getCalls().length, 5);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        assert.deepEqual(existsSync.getCall(1).args, [prestart.mountedConfigFile]);
        assert.deepEqual(existsSync.getCall(2).args, [prestart.vcsConfigFile]);
        assert.deepEqual(existsSync.getCall(3).args, [prestart.defaultProjectConfigFile]);
        assert.deepEqual(existsSync.getCall(4).args, [prestart.configMapConfigFile]);
        sinon.assert.notCalled(writeFileSync);
      });

      it("config exists via env var", function() {
        existsSync.callsFake((...args) => {
          return args[0] === process.env.BRIGADE_CONFIG;
        });

        prestart.createConfig();

        assert.equal(existsSync.getCalls().length, 1);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        sinon.assert.called(writeFileSync);
        sinon.assert.calledWithExactly(writeFileSync, prestart.configFile, "{}");
      });

      it("config exists via mounted file", function() {
        existsSync.callsFake((...args) => {
          return args[0] === prestart.mountedConfigFile;
        });

        prestart.createConfig();

        assert.equal(existsSync.getCalls().length, 2);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        assert.deepEqual(existsSync.getCall(1).args, [prestart.mountedConfigFile]);
        sinon.assert.called(writeFileSync);
        sinon.assert.calledWithExactly(writeFileSync, prestart.configFile, "{}");
      });

      it("no brigade.json mounted, but exists in vcs", function() {
        existsSync.callsFake((...args) => {
          return args[0] === prestart.vcsConfigFile;
        });

        prestart.createConfig();

        assert.equal(existsSync.getCalls().length, 3);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        assert.deepEqual(existsSync.getCall(1).args, [prestart.mountedConfigFile]);
        assert.deepEqual(existsSync.getCall(2).args, [prestart.vcsConfigFile]);
        sinon.assert.called(writeFileSync);
        sinon.assert.calledWithExactly(writeFileSync, prestart.configFile, "{}");
      });

      it("config exists via project default", function() {
        existsSync.callsFake((...args) => {
          return args[0] === prestart.defaultProjectConfigFile;
        });

        prestart.createConfig();

        assert.equal(existsSync.getCalls().length, 4);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        assert.deepEqual(existsSync.getCall(1).args, [prestart.mountedConfigFile]);
        assert.deepEqual(existsSync.getCall(2).args, [prestart.vcsConfigFile]);
        assert.deepEqual(existsSync.getCall(3).args, [prestart.defaultProjectConfigFile]);
        sinon.assert.called(writeFileSync);
        sinon.assert.calledWithExactly(writeFileSync, prestart.configFile, "{}");
      });

      it("config exists via config map", function() {
        existsSync.callsFake((...args) => {
          return args[0] === prestart.configMapConfigFile;
        });

        prestart.createConfig();

        assert.equal(existsSync.getCalls().length, 5);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        assert.deepEqual(existsSync.getCall(1).args, [prestart.mountedConfigFile]);
        assert.deepEqual(existsSync.getCall(2).args, [prestart.vcsConfigFile]);
        assert.deepEqual(existsSync.getCall(3).args, [prestart.defaultProjectConfigFile]);
        assert.deepEqual(existsSync.getCall(4).args, [prestart.configMapConfigFile]);
        sinon.assert.called(writeFileSync);
        sinon.assert.calledWithExactly(writeFileSync, prestart.configFile, "{}");
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
        mock("fs", { existsSync, readFileSync, writeFileSync });

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

        assert.equal(existsSync.getCalls().length, 6);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        assert.deepEqual(existsSync.getCall(1).args, [prestart.mountedConfigFile]);
        assert.deepEqual(existsSync.getCall(2).args, [prestart.vcsConfigFile]);
        assert.deepEqual(existsSync.getCall(3).args, [prestart.defaultProjectConfigFile]);
        assert.deepEqual(existsSync.getCall(4).args, [prestart.configMapConfigFile]);
        assert.deepEqual(existsSync.getCall(5).args, [prestart.configFile]);
        sinon.assert.notCalled(execFileSync);
        sinon.assert.notCalled(exit);
      });

      it("no dependencies object", function() {
        mock(prestart.configFile, {});
        existsSync.callsFake(() => true);

        prestart.addDeps();

        assert.equal(existsSync.getCalls().length, 2);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        assert.deepEqual(existsSync.getCall(1).args, [prestart.configFile]);
        sinon.assert.notCalled(execFileSync);
        sinon.assert.notCalled(exit);
      });

      it("empty dependencies", function() {
        mock(prestart.configFile, { dependencies: {}})
        existsSync.callsFake(() => true);

        prestart.addDeps();

        assert.equal(existsSync.getCalls().length, 2);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        assert.deepEqual(existsSync.getCall(1).args, [prestart.configFile]);
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

        assert.equal(existsSync.getCalls().length, 2);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        assert.deepEqual(existsSync.getCall(1).args, [prestart.configFile]);
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

        assert.equal(existsSync.getCalls().length, 2);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        assert.deepEqual(existsSync.getCall(1).args, [prestart.configFile]);
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

        assert.equal(existsSync.getCalls().length, 2);
        assert.deepEqual(existsSync.getCall(0).args, [process.env.BRIGADE_CONFIG]);
        assert.deepEqual(existsSync.getCall(1).args, [prestart.configFile]);
        sinon.assert.calledOnce(execFileSync);
        sinon.assert.calledWithExactly(execFileSync, "yarn", ["add", "is-thirteen@2.0.0"]);
        sinon.assert.calledOnce(exit);
        sinon.assert.calledWithExactly(exit, 1);
      });
  });
});

