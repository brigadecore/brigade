"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const ts = require("typescript");
const _ts = require("../../ts-internal");
const components_1 = require("../components");
const ERROR_UNSUPPORTED_FILE_ENCODING = -2147024809;
class CompilerHost extends components_1.ConverterComponent {
    getSourceFile(filename, languageVersion, onError) {
        let text;
        try {
            text = ts.sys.readFile(filename, this.application.options.getCompilerOptions().charset);
        }
        catch (e) {
            if (onError) {
                onError(e.number === ERROR_UNSUPPORTED_FILE_ENCODING ? 'Unsupported file encoding' : e.message);
            }
            text = '';
        }
        return text !== undefined ? ts.createSourceFile(filename, text, languageVersion) : undefined;
    }
    getDefaultLibFileName(options) {
        const libLocation = _ts.getDirectoryPath(_ts.normalizePath(ts.sys.getExecutingFilePath()));
        return _ts.combinePaths(libLocation, ts.getDefaultLibFileName(options));
    }
    getDirectories(path) {
        return ts.sys.getDirectories(path);
    }
    getCurrentDirectory() {
        return this.currentDirectory || (this.currentDirectory = ts.sys.getCurrentDirectory());
    }
    useCaseSensitiveFileNames() {
        return ts.sys.useCaseSensitiveFileNames;
    }
    fileExists(fileName) {
        return ts.sys.fileExists(fileName);
    }
    directoryExists(directoryName) {
        return ts.sys.directoryExists(directoryName);
    }
    readFile(fileName) {
        return ts.sys.readFile(fileName);
    }
    getCanonicalFileName(fileName) {
        return ts.sys.useCaseSensitiveFileNames ? fileName : fileName.toLowerCase();
    }
    getNewLine() {
        return ts.sys.newLine;
    }
    writeFile(fileName, data, writeByteOrderMark, onError) { }
}
exports.CompilerHost = CompilerHost;
//# sourceMappingURL=compiler-host.js.map