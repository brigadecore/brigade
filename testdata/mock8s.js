// Store mock pod definitions here as "name" : $resource
mockPods = {
}

// Store configmaps here as "name" : $resource
mockConfigMaps = {
}

kubernetes = {
  withNS: function(ns) {
    return mockCore
  }
}

mockCore = {
  coreV1: {
    pod: {
      get: function(name) {
        return _.find(mockPods, function(ele) {
          return ele.metadata.name == name
        })
      }
      create: function(def) {
        // Succeeded, Running, and Failed are some valid values.
        def.status = { phase: "Succeeded" }
        mockPods[def.metadata.labels.jobname] = def
        return def
      }
    }
  },
  extensions: {
    configmap: {
      create: function(def) {
        mockConfigMaps[def.metadata.labels.jobname] = def
        return def
      }
    }
  }
}
