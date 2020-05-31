

import {HttpClient} from '@angular/common/http';
import {
  Attribute,
  Directive,
  forwardRef,
  Injector,
  Input,
  OnChanges,
  SimpleChanges,
} from '@angular/core';
import {
  AbstractControl,
  AsyncValidator,
  FormControl,
  NG_ASYNC_VALIDATORS,
  NgModel,
  Validator,
} from '@angular/forms';
import {Observable} from 'rxjs/Observable';
import {debounceTime, map} from 'rxjs/operators';

export const uniqueNameValidationKey = 'validImageReference';

/**
 * A validator directive which checks the underlining ngModel's given name is unique or not.
 * If the name exists, error with name `uniqueName` will be added to errors.
 */
@Directive({
  selector: '[kdValidImageReference]',
  providers: [
    {
      provide: NG_ASYNC_VALIDATORS,
      useExisting: forwardRef(() => ValidImageReferenceValidator),
      multi: true,
    },
  ],
})
export class ValidImageReferenceValidator implements AsyncValidator, Validator {
  @Input() namespace: string;

  constructor(private readonly http: HttpClient) {}

  validate(control: AbstractControl): Observable<{[key: string]: string}> {
    if (!control.value) {
      return Observable.of(null);
    } else {
      return this.http
        .post<{valid: boolean; reason: string}>('api/v1/appdeployment/validate/imagereference', {
          reference: control.value,
        })
        .pipe(
          debounceTime(500),
          map(res => (!res.valid ? {[uniqueNameValidationKey]: res.reason} : null)),
        );
    }
  }
}
