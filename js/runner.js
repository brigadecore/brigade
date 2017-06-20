// This is the runner wrapping script.
console.log("Loading ACID core")

function fireEvent(e) {
  var eventHandler = new EventHandler()

  // This goes in the scope in which f() is executed so that we can access
  // the original event without requiring (or, really, allowing) the user to
  // pass the event around.
  exports._event = e

  if (!registerEvents) {
    console.log("no event handlers defined")

    return
  }

  registerEvents(eventHandler)
  console.log("events loaded. Firing " + e.type)

  if (!eventHandler[e.type]) {
    console.log("no event handler registered for " + e.type)

    return
  }

  var f = eventHandler[e.type]

  f(e)
}

exports.fireEvent = fireEvent
