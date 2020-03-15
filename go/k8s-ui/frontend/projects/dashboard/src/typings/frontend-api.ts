



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
