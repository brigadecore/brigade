export interface Worker {
  apiAddress: string
  apiToken: string
  configFilesDirectory: string
  defaultConfigFiles: Map<string, string>
  logLevel?: string
}
