import "mocha";
import { assert } from "chai";
import * as req from "../src/require";

describe("overriding require", function () {
    describe("when `@azure/brigadier` is imported", function () {
        it("correctly overrides to `./brigadier`", function () {
            assert.equal(req.getOverriddenPackage("@azure/brigadier"), "./brigadier");
        });
    });

    describe("when `brigade` is imported", function () {
        it("correctly overrides to `./brigadier`", function () {
            assert.equal(req.getOverriddenPackage("brigade"), "./brigadier");
        });
    });

    describe("when `brigadier` is imported", function () {
        it("correctly overrides to `./brigadier`", function () {
            assert.equal(req.getOverriddenPackage("brigadier"), "./brigadier");
        });
    });

    describe("when a local module relative to the repo is imported", function () {
        it("correctly overrides to the absolute path in the worker pod", function () {
            assert.equal(req.getOverriddenPackage("./local-dir/module"), "/vcs/local-dir/module");
        });
    });

    describe("when global package is imported", function () {
        it("does not affect the import path", function () {
            assert.equal(req.getOverriddenPackage("is-thirteen"), "is-thirteen");
        });
    });
})
