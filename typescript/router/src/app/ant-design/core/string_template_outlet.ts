import {Directive, Input, TemplateRef} from '@angular/core';


@Directive({
  selector: '[stringTemplateOutlet]'
})
export class StringTemplateOutlet {
  @Input()
  set stringTemplateOutlet(value: string | TemplateRef<void>) {

  }
}
