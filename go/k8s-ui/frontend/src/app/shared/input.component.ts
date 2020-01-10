

import {Component, EventEmitter, Input, OnInit, Output, ViewChild} from '@angular/core';
import {ControlValueAccessor} from '@angular/forms';

@Component({
  selector: 'app-input',
  template: `
    <input [(ngModel)]="value" #input [placeholder]="placeholder" [readonly]="readOnly" [style.cursor]="cursor"
           [type]="type" style="padding-left: 6px;" (focus)="focusState = true;" (blur)="focusState = false;" (input)="inputEvent($event)"
           (change)="changeEvent($event)" [style.padding-right.px]="showSearch ? 34 : 6">
    <svg *ngIf="showSearch" width="16" height="16"
         [ngStyle]="{'fill': focusState ? '#222' : '#666', 'stroke': focusState ? '#222' : '#666'}"
         xmlns="http://www.w3.org/2000/svg" viewBox="0, 0, 100, 100">
      <circle cx="40" cy="45" r="35" stroke-width="10" fill="#fff"></circle>
      <line x1="70" y1="65" x2="95" y2="85" stroke-width="10"></line>
      <circle cx="95" cy="85" r="4"></circle>
    </svg>
  `,
  styleUrls: ['./input.component.scss']
})

export class InputComponent implements ControlValueAccessor {

  constructor() {
  }

  @Input()
  set search(value) {
    if (value !== undefined) {
      this.showSearch = true;
    }
  }
  @ViewChild('input', { static: false }) inputElement;
  @Input() placeholder = '';
  @Input() type = 'text';
  @Input() cursor = 'auto';
  value: string;
  readOnly = false;
  showSearch = false;
  focusState = false;

  @Output() changed = new EventEmitter<any>();

  registerOnChange(fn: any): void {
  }

  registerOnTouched(fn: any): void {
  }

  setDisabledState(isDisabled: boolean): void {
  }

  writeValue(obj: any): void {
  }

  changeEvent(event: any) {
    this.changed.emit(event);
  }
}
