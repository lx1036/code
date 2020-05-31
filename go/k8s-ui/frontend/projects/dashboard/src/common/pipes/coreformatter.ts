

import {DecimalPipe} from '@angular/common';
import {Pipe} from '@angular/core';

/**
 * Formats cores usage in millicores to a decimal prefix format, e.g. 321,20 kCPU.
 */
@Pipe({name: 'kdCores'})
export class CoreFormatter extends DecimalPipe {
  readonly base = 1000;
  readonly powerSuffixes = ['m', '', 'k', 'M', 'G', 'T'];

  transform(value: number): string {
    let divider = 1;
    let power = 0;

    while (value / divider > this.base && power < this.powerSuffixes.length - 1) {
      divider *= this.base;
      power += 1;
    }

    const formatted = super.transform(value / divider, '1.2-2');
    const suffix = this.powerSuffixes[power];
    return suffix ? `${formatted}${suffix}` : formatted;
  }
}
