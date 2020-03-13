



export interface KdError {
  status: string;
  code: number;
  message: string;

  localize(): KdError;
}


export interface StateError {
  error: KdError;
}
