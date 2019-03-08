const {Job} = require("./brigadier");

exports.alpineJob = function(name) {
  j = new Job(name, "alpine:3.7", ["echo hello"])
  return j
}
