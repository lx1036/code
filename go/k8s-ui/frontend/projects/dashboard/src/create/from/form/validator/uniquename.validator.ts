

import {HttpClient} from '@angular/common/http';
import {Directive, forwardRef, Input} from '@angular/core';
import {
  AbstractControl,
  AsyncValidator,
  AsyncValidatorFn,
  NG_ASYNC_VALIDATORS,
} from '@angular/forms';
import {Observable} from 'rxjs/Observable';
import {debounceTime, map} from 'rxjs/operators';

export const uniqueNameValidationKey = 'uniqueName';

/**
 * A validator directive which checks the underlining ngModel's given name is unique or not.
 * If the name exists, error with name `uniqueName` will be added to errors.
 */
@Directive({
  selector: '[kdUniqueName]',
  providers: [
    {
      provide: NG_ASYNC_VALIDATORS,
      useExisting: forwardRef(() => UniqueNameValidator),
      multi: true,
    },
  ],
})
export class UniqueNameValidator implements AsyncValidator {
  @Input() namespace: string;

  constructor(private readonly http: HttpClient) {}

  validate(control: AbstractControl): Observable<{[key: string]: boolean} | null> {
    return validateUniqueName(this.http, this.namespace)(control) as Observable<{
      [key: string]: boolean;
    } | null>;
  }
}

export function validateUniqueName(http: HttpClient, namespace: string): AsyncValidatorFn {
  return (control: AbstractControl): Observable<{[key: string]: boolean} | null> => {
    if (!control.value) {
      return Observable.of(null);
    } else {
      return http
        .post<{valid: boolean}>('api/v1/appdeployment/validate/name', {
          name: control.value,
          namespace,
        })
        .pipe(
          debounceTime(500),
          map(res => (!res.valid ? {[uniqueNameValidationKey]: control.value} : null)),
        );
    }
  };
}
