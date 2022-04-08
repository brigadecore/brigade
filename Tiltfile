load('ext://min_k8s_version', 'min_k8s_version')
min_k8s_version('1.18.0')

trigger_mode(TRIGGER_MODE_MANUAL)

load('ext://namespace', 'namespace_create')
namespace_create('brigade')
k8s_resource(
  new_name = 'namespace',
  objects = ['brigade:namespace'],
  labels = ['brigade']
)

config.clear_enabled_resources()
config.set_enabled_resources([
  'apiserver',
  'artemis',
  'logger',
  'mongodb',
  'namespace',
  'observer',
  'scheduler'
])

update_settings(
  suppress_unused_image_warnings = ["brigadecore/brigade2-git-initializer"]
)
update_settings(
  suppress_unused_image_warnings = ["brigadecore/brigade2-worker"]
)

docker_build(
  'brigadecore/brigade2-apiserver', '.',
  dockerfile = 'v2/apiserver/Dockerfile',
  only = [
    'sdk/',
    'v2/apiserver/',
    'v2/internal/',
    'v2/go.mod',
    'v2/go.sum'
  ],
  ignore = ['**/*_test.go']
)
k8s_resource(
  workload = 'brigade-apiserver',
  new_name = 'apiserver',
  resource_deps = ['artemis', 'mongodb'],
  port_forwards = '31600:8080',
  labels = ['brigade']
)
k8s_resource(
  workload = 'apiserver',
  objects = [
    'brigade-apiserver:clusterrole',
    'brigade-apiserver:clusterrolebinding',
    'brigade-apiserver:secret',
    'brigade-apiserver:serviceaccount'
  ]
)

docker_build(
  'brigadecore/brigade2-artemis', '.',
  dockerfile = 'v2/artemis/Dockerfile',
  only = ['v2/artemis/'],
  ignore = ['**/*_test.go']
)
k8s_resource(
  workload = 'brigade-artemis',
  new_name = 'artemis',
  labels = ['brigade']
)
k8s_resource(
  workload = 'artemis',
  objects = [
    'brigade-artemis:configmap',
    'brigade-artemis:secret',
    'brigade-artemis-common-config:secret'
  ]
)

docker_build(
  'brigadecore/brigade2-git-initializer', '.',
  dockerfile = 'v2/git-initializer/Dockerfile',
  only = [
    'sdk/',
    'v2/git-initializer/',
    'v2/internal/',
    'v2/go.mod',
    'v2/go.sum'
  ],
  ignore = ['**/*_test.go'],
  match_in_env_vars = True
)

docker_build(
  'brigadecore/brigade2-logger', '.',
  dockerfile = 'v2/logger/Dockerfile',
  only = ['v2/logger/'],
  ignore = ['**/*_test.go']
)
k8s_resource(
  workload = 'brigade-logger',
  new_name = 'logger',
  labels = ['brigade'],
)
k8s_resource(
  workload = 'logger',
  objects = [
    'brigade-logger:clusterrole',
    'brigade-logger:clusterrolebinding',
    'brigade-logger:secret',
    'brigade-logger:serviceaccount'
  ]
)
k8s_resource(
  workload = 'brigade-logger-windows',
  new_name = 'logger-windows',
  labels = ['brigade'],
)

k8s_resource(
  workload = 'brigade-mongodb',
  new_name = 'mongodb',
  labels = ['brigade']
)
k8s_resource(
  workload = 'mongodb',
  objects = [
    'brigade-mongodb:persistentvolumeclaim',
    'brigade-mongodb:secret',
    'brigade-mongodb:serviceaccount'
  ]
)

docker_build(
  'brigadecore/brigade2-observer', '.',
  dockerfile = 'v2/observer/Dockerfile',
  only = [
    'sdk/',
    'v2/internal/',
    'v2/observer/',
    'v2/go.mod',
    'v2/go.sum'
  ],
  ignore = ['**/*_test.go']
)
k8s_resource(
  workload = 'brigade-observer',
  new_name = 'observer',
  resource_deps = ['apiserver'],
  labels = ['brigade']
)
k8s_resource(
  workload = 'observer',
  objects = [
    'brigade-observer:clusterrole',
    'brigade-observer:clusterrolebinding',
    'brigade-observer:secret',
    'brigade-observer:serviceaccount'
  ]
)

docker_build(
  'brigadecore/brigade2-scheduler',
  '.',
  dockerfile = 'v2/scheduler/Dockerfile',
  only = [
    'sdk/',
    'v2/internal/',
    'v2/scheduler/',
    'v2/go.mod',
    'v2/go.sum'
  ],
  ignore = ['**/*_test.go']
)
k8s_resource(
  workload = 'brigade-scheduler',
  new_name = 'scheduler',
  resource_deps = ['apiserver', 'artemis'],
  labels = ['brigade'],
)
k8s_resource(
  workload = 'scheduler',
  objects = [
    'brigade-scheduler:secret',
    'brigade-scheduler:serviceaccount',
  ]
)

docker_build(
  'brigadecore/brigade2-worker', '.',
  dockerfile = 'v2/worker/Dockerfile',
  only = [
    'v2/brigadier/',
    'v2/brigadier-polyfill/',
    'v2/worker'
  ],
  ignore = ['**/*_test.go'],
  match_in_env_vars = True
)

k8s_yaml(
  helm(
    './charts/brigade',
    name = 'brigade',
    namespace = 'brigade',
    set = [
      'apiserver.rootUser.password=F00Bar!!!',
      'apiserver.tls.enabled=false',
      'artemis.password=insecure-artemis-password',
      'gitInitializer.linux.image.repository=brigadecore/brigade2-git-initializer',
      'observer.apiToken=insecure-observer-token',
      'scheduler.apiToken=insecure-scheduler-token',
      'worker.image.repository=brigadecore/brigade2-worker'
    ],
  ),
)
