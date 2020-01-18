import {Component, Input, OnInit} from '@angular/core';

@Component({
  selector: 'app-card',
  template: `
    <div *ngIf="cardTitle" class="card-title">{{cardTitle}}</div>
    <ng-content></ng-content>
  `
})

export class CardComponent {
  cardTitle: string;

  @Input()
  set header(value: string) {
    if (value) {
      this.cardTitle = value;
    }
  }
}
