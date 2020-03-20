/**
 * Frontend specific errors or errors transformed based on server response.
 */
import {HttpErrorResponse} from "@angular/common/http";
import {KdError as KdApiError} from "../../typings/frontend-api";
import {ErrStatus, K8SError as K8SApiError} from "../../typings/backend-api";


export enum ApiError {
  tokenExpired = 'MSG_TOKEN_EXPIRED_ERROR',
  encryptionKeyChanged = 'MSG_ENCRYPTION_KEY_CHANGED',
}

export enum ErrorStatus {
  unauthorized = 'Unauthorized',
  forbidden = 'Forbidden',
  internal = 'Internal error',
  unknown = 'Unknown error',
}

export enum ErrorCode {
  unauthorized = 401,
  forbidden = 403,
  internal = 500,
}

const localizedErrors: {[key: string]: string} = {
  MSG_TOKEN_EXPIRED_ERROR: 'You have been logged out because your token has expired.',
  MSG_ENCRYPTION_KEY_CHANGED: 'You have been logged out because your token is invalid.',
  MSG_ACCESS_DENIED: 'Access denied.',
  MSG_DASHBOARD_EXCLUSIVE_RESOURCE_ERROR: 'Trying to access/modify dashboard exclusive resource.',
  MSG_LOGIN_UNAUTHORIZED_ERROR: 'Invalid credentials provided',
};

export class K8SError implements K8SApiError {
  ErrStatus: ErrStatus;

  constructor(error: ErrStatus) {
    this.ErrStatus = error;
  }

  toKdError(): KdError {
    return new KdError(this.ErrStatus.reason, this.ErrStatus.code, this.ErrStatus.message);
  }
}


export class KdError implements KdApiError {
  constructor(public status: string, public code: number, public message: string) {}

  static isError(error: HttpErrorResponse, ...apiErrors: string[]): boolean {
    // API errors will set 'error' as a string.
    if (typeof error.error === 'object') {
      return false;
    }

    for (const apiErr of apiErrors) {
      if (apiErr === (error.error as string).trim()) {
        return true;
      }
    }

    return false;
  }

  localize(): KdError {
    const result = this;

    const localizedErr = localizedErrors[this.message.trim()];
    if (localizedErr) {
      this.message = localizedErr;
    }

    return result;
  }
}

export function AsKdError(error: HttpErrorResponse): KdError {
  const result = {} as KdError;
  let status: string;

  result.message = error.message;
  result.code = error.status;

  if (typeof error.error !== 'object') {
    result.message = error.error;
  }

  switch (error.status) {
    case ErrorCode.unauthorized:
      status = ErrorStatus.unauthorized;
      break;
    case ErrorCode.forbidden:
      status = ErrorStatus.forbidden;
      break;
    case ErrorCode.internal:
      status = ErrorStatus.internal;
      break;
    default:
      status = ErrorStatus.unknown;
  }

  result.status = status;
  return new KdError(result.status, result.code, result.message).localize();
}

export const ERRORS = {
  forbidden: new KdError(
    ErrorStatus.forbidden,
    ErrorCode.forbidden,
    localizedErrors.MSG_ACCESS_DENIED,
  ),
};

