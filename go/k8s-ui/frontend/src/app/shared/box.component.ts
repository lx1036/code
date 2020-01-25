import {Component, HostListener, OnInit} from '@angular/core';

@Component({
  selector: 'app-box',
  template: `
    <ng-content></ng-content>
  `,
  styleUrls: ['./box.component.scss']
})
export class BoxComponent {
  enter: boolean;



  @HostListener('mouseenter')
  enterEvent() {
    this.enter = true;
  }

  @HostListener('mouseleave')
  leaveEvent() {
    this.enter = false;
  }
}
