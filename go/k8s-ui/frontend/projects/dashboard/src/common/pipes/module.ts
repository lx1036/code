

import {NgModule} from '@angular/core';

import {CommaSeparatedPipe} from './commaseparated';
import {CoreFormatter} from './coreformatter';
import {MemoryFormatter} from './memoryformatter';
import {RelativeTimeFormatter} from './relativetime';
import {SafeHtmlFormatter} from './safehtml';
import {TrimPipe} from './trim';

@NgModule({
  declarations: [
    MemoryFormatter,
    CoreFormatter,
    RelativeTimeFormatter,
    SafeHtmlFormatter,
    TrimPipe,
    CommaSeparatedPipe,
  ],
  exports: [
    MemoryFormatter,
    CoreFormatter,
    RelativeTimeFormatter,
    SafeHtmlFormatter,
    TrimPipe,
    CommaSeparatedPipe,
  ],
})
export class PipesModule {}
