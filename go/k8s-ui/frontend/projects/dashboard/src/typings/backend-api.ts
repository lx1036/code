import {KdError} from "./frontend-api";


export type AuthenticationMode = string;

export interface EnabledAuthenticationModes {
  modes: AuthenticationMode[];
}

export interface LoginSkippableResponse {
  skippable: boolean;
}

export interface LoginSpec {
  username: string;
  password: string;
  token: string;
  kubeConfig: string;
}

export interface K8sError {
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
  errors: K8sError[];
}
