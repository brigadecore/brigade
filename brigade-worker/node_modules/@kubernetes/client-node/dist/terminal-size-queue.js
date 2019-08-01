"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const stream_1 = require("stream");
class TerminalSizeQueue extends stream_1.Readable {
    constructor(opts = {}) {
        super(Object.assign({}, opts, { 
            // tslint:disable-next-line:no-empty
            read() { } }));
    }
    handleResizes(writeStream) {
        // Set initial size
        this.resize(getTerminalSize(writeStream));
        // Handle future size updates
        writeStream.on('resize', () => this.resize(getTerminalSize(writeStream)));
    }
    resize(size) {
        this.push(JSON.stringify(size));
    }
}
exports.TerminalSizeQueue = TerminalSizeQueue;
function isResizable(stream) {
    if (stream == null) {
        return false;
    }
    const hasRows = 'rows' in stream;
    const hasColumns = 'columns' in stream;
    const hasOn = typeof stream.on === 'function';
    return hasRows && hasColumns && hasOn;
}
exports.isResizable = isResizable;
function getTerminalSize(writeStream) {
    return { height: writeStream.rows, width: writeStream.columns };
}
//# sourceMappingURL=terminal-size-queue.js.map