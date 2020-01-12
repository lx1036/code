

import {Component, EventEmitter, forwardRef, Input, OnInit, Output, ViewChild} from '@angular/core';
import {ControlValueAccessor, NG_VALUE_ACCESSOR} from '@angular/forms';

@Component({
  selector: 'app-input',
  template: `
    <input [(ngModel)]="value" #input [placeholder]="placeholder" [readonly]="readOnly" [style.cursor]="cursor"
           [type]="type" style="padding-left: 6px;" (focus)="focusState = true;" (blur)="focusState = false;" (input)="inputEvent($event.target.value)"
           (change)="changeEvent($event)" [style.padding-right.px]="showSearch ? 34 : 6">
    <svg *ngIf="showSearch" width="16" height="16" [ngStyle]="{'fill': focusState ? '#222' : '#666', 'stroke': focusState ? '#222' : '#666'}"
         xmlns="http://www.w3.org/2000/svg" viewBox="0, 0, 100, 100">
      <circle cx="40" cy="45" r="35" stroke-width="10" fill="#fff"></circle>
      <line x1="70" y1="65" x2="95" y2="85" stroke-width="10"></line>
      <circle cx="95" cy="85" r="4"></circle>
    </svg>
  `,
  styleUrls: ['./input.component.scss'],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => InputComponent),
      multi: true,
    },
  ]
})

export class InputComponent implements ControlValueAccessor {
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
  showSearch = true;
  focusState = false;

  @Output() changed = new EventEmitter<any>();

  fn: (_) => {};

  registerOnChange(fn: (_) => {}): void {
    this.fn = fn;
  }
  inputEvent(value) {
    this.fn(value);
  }

  registerOnTouched(fn: () => {}): void {
  }

  setDisabledState(isDisabled: boolean): void {
  }

  writeValue(value: string): void {
    if (this.value !== value) {
      this.value = value;
    }
  }

  changeEvent(event: any) {
    this.changed.emit(event);
  }
}
