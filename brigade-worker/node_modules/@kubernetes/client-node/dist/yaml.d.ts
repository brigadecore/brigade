import * as yaml from 'js-yaml';
export declare function loadYaml<T>(data: string, opts?: yaml.LoadOptions): T;
export declare function loadAllYaml(data: string, opts?: yaml.LoadOptions): any[];
export declare function dumpYaml(object: any, opts?: yaml.DumpOptions): string;
