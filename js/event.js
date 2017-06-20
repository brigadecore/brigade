// EventHandler describes the list of events that Acid is aware of.
function EventHandler() {
  // Every event handler gets the param 'data', which is the body of the request.
  this.push = function() {}
  this.pullRequest = function() {}
}

exports.EventHandler = EventHandler
