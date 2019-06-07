# Javascript Kubernetes Client information

[![Build Status](https://travis-ci.org/kubernetes-client/javascript.svg?branch=master)](https://travis-ci.org/kubernetes-client/javascript)
[![Client Capabilities](https://img.shields.io/badge/Kubernetes%20client-Silver-blue.svg?style=flat&colorB=C0C0C0&colorA=306CE8)](http://bit.ly/kubernetes-client-capabilities-badge)
[![Client Support Level](https://img.shields.io/badge/kubernetes%20client-beta-green.svg?style=flat&colorA=306CE8)](http://bit.ly/kubernetes-client-support-badge)


The Javascript clients for Kubernetes is implemented in
[typescript](https://typescriptlang.org), but can be called from either
Javascript or Typescript.

For now, the client is implemented for server-side use with node
using the `request` library.

There are future plans to also build a jQuery compatible library but
for now, all of the examples and instructions assume the node client.

# Installation
```sh
$ npm install @kubernetes/client-node
```

# Example code

## List all pods
```javascript
const k8s = require('@kubernetes/client-node');


const kc = new k8s.KubeConfig();
kc.loadFromDefault();

const k8sApi = kc.makeApiClient(k8s.Core_v1Api);

k8sApi.listNamespacedPod('default')
    .then((res) => {
        console.log(res.body);
    });
```

## Create a new namespace
```javascript
const k8s = require('@kubernetes/client-node');

const kc = new k8s.KubeConfig();
kc.loadFromDefault();

const k8sApi = kc.makeApiClient(k8s.Core_v1Api);

var namespace = {
  metadata: {
    name: 'test'
  }
};

k8sApi.createNamespace(namespace).then(
  (response) => {
    console.log('Created namespace');
    console.log(response);
    k8sApi.readNamespace(namespace.metadata.name).then(
      (response) => {
        console.log(response);
        k8sApi.deleteNamespace(
          namespace.metadata.name, {} /* delete options */);
      });
  },
  (err) => {
    console.log('Error!: ' + err);
  }
);
```

# Additional Examples
There are several more examples in the [examples](https://github.com/kubernetes-client/javascript/tree/master/examples) directory.

# Development

All dependencies of this project are expressed in its
[`package.json` file](package.json). Before you start developing, ensure
that you have [NPM](https://www.npmjs.com/) installed, then run:

```console
npm install
```

## (re) Generating code

```
cd ../
git clone https://github.com/kubernetes-client/gen
cd javascript
../gen/openapi/typescript.sh src settings
```

# Testing

Tests are written using the [Chai](http://chaijs.com/) library. See
[`config_test.ts`](./config_test.ts) for an example.

To run tests, execute the following:

```console
npm test
```

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to contribute.
