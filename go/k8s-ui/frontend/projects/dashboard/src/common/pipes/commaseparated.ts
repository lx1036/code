

import {Pipe, PipeTransform} from '@angular/core';

@Pipe({name: 'commaSeparated'})
export class CommaSeparatedPipe implements PipeTransform {
  transform(value: string[]): string {
    if (!value) {
      return '';
    }
    return value.join(', ');
  }
}
