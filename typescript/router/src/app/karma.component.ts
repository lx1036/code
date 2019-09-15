import { Component } from '@angular/core';

@Component({
  selector: 'karma-root',
  template: `<p>Hello World {{title}}!</p>`,
})
export class KarmaComponent {
  title = 'angular';
}
