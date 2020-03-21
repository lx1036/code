import {KdError} from "./frontend-api";


export interface LoginSpec {
  username: string;
  password: string;
  token: string;
  kubeConfig: string;
}

export interface LoginStatus {
  tokenPresent: boolean;
  headerPresent: boolean;
  httpsMode: boolean;
}

export interface K8SError {
  ErrStatus: ErrStatus;

  toKdError(): KdError;
}

export interface ErrStatus {
  message: string;
  code: number;
  status: string;
  reason: string;
}

export interface CsrfToken {
  token: string;
}

export interface AuthResponse {
  jweToken: string;
  errors: K8SError[];
}
