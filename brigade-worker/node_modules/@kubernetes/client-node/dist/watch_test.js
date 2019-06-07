"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const chai_1 = require("chai");
const ts_mockito_1 = require("ts-mockito");
const config_1 = require("./config");
const watch_1 = require("./watch");
describe('Watch', () => {
    it('should construct correctly', () => {
        const kc = new config_1.KubeConfig();
        const watch = new watch_1.Watch(kc);
    });
    it('should watch correctly', () => {
        const kc = new config_1.KubeConfig();
        const server = 'foo.company.com';
        kc.clusters = [
            {
                name: 'cluster',
                server,
            },
        ];
        kc.contexts = [
            {
                cluster: 'cluster',
                user: 'user',
            },
        ];
        kc.users = [
            {
                name: 'user',
            },
        ];
        const fakeRequestor = ts_mockito_1.mock(watch_1.DefaultRequest);
        const watch = new watch_1.Watch(kc, ts_mockito_1.instance(fakeRequestor));
        const obj1 = {
            type: 'ADDED',
            object: {
                foo: 'bar',
            },
        };
        const obj2 = {
            type: 'MODIFIED',
            object: {
                baz: 'blah',
            },
        };
        const fakeRequest = {
            pipe: (stream) => {
                stream.write(JSON.stringify(obj1) + '\n');
                stream.write(JSON.stringify(obj2) + '\n');
            },
        };
        ts_mockito_1.when(fakeRequestor.webRequest(ts_mockito_1.anything(), ts_mockito_1.anyFunction())).thenReturn(fakeRequest);
        const path = '/some/path/to/object';
        const receivedTypes = [];
        const receivedObjects = [];
        let doneCalled = false;
        let doneErr;
        watch.watch(path, {}, (phase, obj) => {
            receivedTypes.push(phase);
            receivedObjects.push(obj);
        }, (err) => {
            doneCalled = true;
            doneErr = err;
        });
        ts_mockito_1.verify(fakeRequestor.webRequest(ts_mockito_1.anything(), ts_mockito_1.anyFunction()));
        const [opts, doneCallback] = ts_mockito_1.capture(fakeRequestor.webRequest).last();
        const reqOpts = opts;
        chai_1.expect(reqOpts.uri).to.equal(`${server}${path}`);
        chai_1.expect(reqOpts.method).to.equal('GET');
        chai_1.expect(reqOpts.json).to.equal(true);
        chai_1.expect(receivedTypes).to.deep.equal([obj1.type, obj2.type]);
        chai_1.expect(receivedObjects).to.deep.equal([obj1.object, obj2.object]);
        chai_1.expect(doneCalled).to.equal(false);
        doneCallback(null, null, null);
        chai_1.expect(doneCalled).to.equal(true);
        chai_1.expect(doneErr).to.equal(null);
        const errIn = { error: 'err' };
        doneCallback(errIn, null, null);
        chai_1.expect(doneErr).to.deep.equal(errIn);
    });
});
//# sourceMappingURL=watch_test.js.map