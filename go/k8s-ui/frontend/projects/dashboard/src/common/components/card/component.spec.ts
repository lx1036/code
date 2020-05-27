

import {Component, CUSTOM_ELEMENTS_SCHEMA, DebugElement} from '@angular/core';
import {async, ComponentFixture, TestBed} from '@angular/core/testing';
import {MatCardModule} from '@angular/material/card';
import {MatDividerModule} from '@angular/material/divider';
import {MatIconModule} from '@angular/material/icon';
import {MatTooltip, MatTooltipModule} from '@angular/material/tooltip';
import {By} from '@angular/platform-browser';
import {NoopAnimationsModule} from '@angular/platform-browser/animations';

import {CardComponent} from './component';

@Component({
  selector: 'test',
  template: `
    <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet" />
    <kd-card [expanded]="isExpanded" [expandable]="isExpandable" role="table">
      <div title>{{ title }}</div>
      <div description>Description: default</div>
      <div actions>Actions: default</div>
      <div content>Content: default</div>
      <div footer>Footer: default</div>
    </kd-card>
  `,
})
class TestComponent {
  title = 'my-card-default-title';
  isExpanded = true;
  isExpandable = true;
}

describe('CardComponent', () => {
  let component: TestComponent;
  let fixture: ComponentFixture<TestComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [CardComponent, TestComponent],
      imports: [
        MatIconModule,
        MatCardModule,
        MatDividerModule,
        MatTooltipModule,
        NoopAnimationsModule,
      ],
      schemas: [CUSTOM_ELEMENTS_SCHEMA],
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TestComponent);
    component = fixture.componentInstance;
  });

  it('shows the title div when withTitle==true', () => {
    const title = 'my-card-default-expanded';

    component.title = title;

    component.isExpanded = true;
    fixture.detectChanges();
    const card = fixture.debugElement.query(By.css('mat-card-title'));
    expect(card).toBeTruthy();
    const content = card.query(By.css('div[content]'));
    expect(content).toBeFalsy();
    const titleNative = card.query(By.css('div[title] ')).nativeElement;
    expect(titleNative.innerHTML).toBe(title);
  });

  it('hides the title div when withTitle==false', () => {
    const title = 'my-card-default-not-expanded';

    component.title = title;
    component.isExpanded = false;
    fixture.detectChanges();
    const card = fixture.debugElement.query(By.css('mat-card-title'));
    expect(card).toBeTruthy();
    const content = card.query(By.css('div[content]'));
    expect(content).toBeFalsy();
    const titleNative = card.query(By.css('div[title] ')).nativeElement;
    expect(titleNative.innerHTML).toBe(title);
  });
});
