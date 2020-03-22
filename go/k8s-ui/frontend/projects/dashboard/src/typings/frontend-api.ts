



export interface KdError {
  status: string;
  code: number;
  message: string;

  localize(): KdError;
}


export interface StateError {
  error: KdError;
}

export interface KdFile {
  name: string;
  content: string;
}

export interface HTMLInputEvent extends Event {
  target: HTMLInputElement & EventTarget;
}


export interface PluginMetadata {
  name: string;
  path: string;
  dependencies: string[];
}

export interface PluginsConfig {
  status: number;
  plugins: PluginMetadata[];
  errors?: object[];
}
