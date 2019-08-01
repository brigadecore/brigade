var util = require('util'),
    TransformStream = require('stream').Transform;

module.exports = function (options) {
  return new JSONStream(options);
};

var JSONStream = module.exports.JSONStream = function (options) {
  options = options || {};
  TransformStream.call(this, options);
  this._writableState.objectMode = false;
  this._readableState.objectMode = true;
  this._async = options.async || false;
};
util.inherits(JSONStream, TransformStream);

JSONStream.prototype._transform = function (data, encoding, callback) {
  if (!Buffer.isBuffer(data)) data = new Buffer(data);
  if (this._buffer) {
    data = Buffer.concat([this._buffer, data]);
  }

  var ptr = 0, start = 0;
  while (++ptr <= data.length) {
    if (data[ptr] === 10 || ptr === data.length) {
      var line;
      try {
        line = JSON.parse(data.slice(start, ptr));
      }
      catch (ex) { }
      if (line) {
        this.push(line);
        line = null;
      }
      if (data[ptr] === 10) start = ++ptr;
    }
  }

  this._buffer = data.slice(start);
  return this._async
    ? void setImmediate(callback)
    : void callback();
};
