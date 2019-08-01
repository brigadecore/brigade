Changelog for the Kubernetes typescript client.

## 0.10.0
  * BREAKING CHANGE: API class names are CamelCase. This replaces the the "lower  case version" naming. For example, `Core_v1Api` is now `CoreV1Api` and `Events_v1beta1Api` is now `EventsV1beta1Api`.

## 0.8.1
  * Fix an issue with exposing bluebird types for `Promise` that broke `es6` users.
