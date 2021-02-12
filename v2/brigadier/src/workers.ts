export interface Worker {
  apiAddress: string
  apiToken: string
  configFilesDirectory: string
  defaultConfigFiles: { [key: string]: string }
  logLevel?: string
}
