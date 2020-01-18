import {Component, HostListener, Input, OnInit} from '@angular/core';






@Component({
  selector: 'app-dropdown-item',
  template: `
    <h4 *ngIf="title">{{title}}</h4>
    <div class="inner">
      <ng-content class="aaa" style="color: red;"></ng-content>
    </div>
  `
})

export class DropdownItemComponent implements OnInit {
  @Input() title: string;

  constructor() {
  }

  ngOnInit() {
  }
}

@Component({
  selector: 'app-dropdown',
  template: `
    <div class="item" [style.opacity]="showContent ? 1 : 0.7">
      <ng-content></ng-content>
    </div>
    <div *ngIf="showContent" @contentState class="content" [class.nowrap]="size === 'small'" [style.width.px]="width"
         [style.right.px]="right" [style.maxHeight.px]="maxHeight" (mouseenter)="enterEvent()" (mouseleave)="leaveEvent()">
      <div class="scrollContent" [style.marginRight.px]="marginRight" [style.maxHeight.px]="maxHeight" (scroll)="scrollEvent($event)">
        <ng-content select="app-dropdown-item"></ng-content>
        <div class="scrollBar" [@barState]="barState" [style.transform]="translateY" [style.height.px]="barStyle.height" (mousedown)="downEvent($event)"></div>
      </div>
    </div>
  `,
  styleUrls: ['./dropdown.component.scss'],
  animations: [

  ]
})

export class DropdownComponent implements OnInit {
  showContent = false;
  // size 默认为空。如果传入small，则是最小自适应，传入middle，为50%宽度。
  @Input() size = '';
  // 这里是处理当item是最接近右边栏时候。采用right定位，防止出现滚动条。
  @Input() last;
  right: number | string = 0;
  width: number | string = 0;
  maxHeight = 400;
  marginRight = 0;
  barState = 'hide';
  barStyle = {
    height: 0,
    top: 0
  };
  get translateY() {
    return `translateY(${this.barStyle.top}%)`;
  }

  constructor() {}

  ngOnInit() {
  }

  @HostListener('mouseenter')
  enterEvent() {

  }

  @HostListener('mouseleave')
  leaveEvent() {
    this.showContent = false;
  }

  scrollEvent(event) {

  }

  downEvent(event) {

  }
}
