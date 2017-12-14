import * as PrettyError from "pretty-error";
import * as path from "path";

const rootPath = path.join(__dirname, "..");
const pe = new PrettyError()
  .skipNodeFiles()
  .skipPackage("ts-node")
  .skipPackage("bluebird")
  .alias(rootPath, ".");
pe.start();

export interface Logger {
  error(message?: any, ...optionalParams: any[]): void;
  log(message?: any, ...optionalParams: any[]): void;
}

export class ContextLogger implements Logger {
  context: string;
  constructor(...ctx: string[]) {
    this.context = `[${new Array("brigade", ...ctx).join(":")}]`;
  }
  error(message?: any, ...optionalParams: any[]): void {
    console.error(this.context, message, ...optionalParams);
  }
  log(message?: any, ...optionalParams: any[]): void {
    console.log(this.context, message, ...optionalParams);
  }
}
