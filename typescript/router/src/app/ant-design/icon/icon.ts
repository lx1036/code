import {Directive, Input} from '@angular/core';
import {InputBoolean} from '../core/decorator';

// svg folder names
export type ThemeType = 'fill' | 'outline' | 'twotone';

@Directive({
  selector: '[ng-icon]'
})
export class IconDirective {
  @Input() type:
  @Input() @InputBoolean() spin;
}
