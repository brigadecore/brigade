/// <reference types="node" />
import { Readable, ReadableOptions } from 'stream';
export interface ResizableStream {
    columns: number;
    rows: number;
    on(event: 'resize', cb: () => void): any;
}
export interface TerminalSize {
    height: number;
    width: number;
}
export declare class TerminalSizeQueue extends Readable {
    constructor(opts?: ReadableOptions);
    handleResizes(writeStream: ResizableStream): void;
    private resize;
}
export declare function isResizable(stream: any): boolean;
