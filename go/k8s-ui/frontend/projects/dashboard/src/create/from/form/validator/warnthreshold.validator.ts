

import {Directive, forwardRef, Input} from '@angular/core';
import {AbstractControl, NG_VALIDATORS, Validator} from '@angular/forms';

/**
 * Description:
 *    Directive can be used to warn user that some number may be too big for given input field
 *    but not block the actual form submit.
 *
 * Params:
 *    - kdWarnThreshold - Max number that is allowed before input field will be marked with
 *      a warning.
 *    - kdWarnThresholdBind - boolean page scope variable that should be bound to directive. Warning
 *      state change will be populated to this variable.
 *
 * Usage:
 *    <md-input-container>
 *      <input type="number" name="maxInput" kd-warn-threshold="100"
 *             kd-warn-threshold-bind="showWarning">
 *      <span class="kd-warn-threshold" ng-show="showWarning">Warning message</span>
 *    </md-input-container>
 *
 */
@Directive({
  selector: '[kdWarnThreshold]',
  providers: [
    {
      provide: NG_VALIDATORS,
      useExisting: forwardRef(() => WarnThresholdValidator),
      multi: true,
    },
  ],
})
export class WarnThresholdValidator implements Validator {
  @Input() kdWarnThreshold: number;
  hasWarning: boolean;

  constructor() {}

  validate(control: AbstractControl): {[key: string]: object} {
    if (this.shouldSetWarning(control.value)) {
      this.hasWarning = true;
      return {warnThreshold: {value: '333'}};
    }

    if (this.shouldRemoveWarning(control.value)) {
      this.hasWarning = false;
      return null;
    }
    return null;
  }

  /**
   * Returns true if input number is larger than max allowed number provided as attribute and
   * there is no warning already set, false otherwise.
   */
  shouldSetWarning(value: number): boolean {
    return value > this.kdWarnThreshold;
  }

  /**
   * Returns true if input number is lower or equal to max allowed number provided as attribute
   * and there is still a warning set, false otherwise.
   */
  shouldRemoveWarning(value: number): boolean {
    return !this.hasWarning && !this.shouldSetWarning(value);
  }
}
