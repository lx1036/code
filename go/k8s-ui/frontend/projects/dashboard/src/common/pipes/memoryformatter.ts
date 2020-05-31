

import {DecimalPipe} from '@angular/common';
import {Pipe} from '@angular/core';

/**
 * Formats memory in bytes to a binary prefix format, e.g., 789,21 MiB.
 */
@Pipe({name: 'kdMemory'})
export class MemoryFormatter extends DecimalPipe {
  readonly base = 1024;
  readonly powerSuffixes = ['', 'Ki', 'Mi', 'Gi', 'Ti', 'Pi'];

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
