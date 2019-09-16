import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import {BrowserTestingModule} from "@angular/platform-browser/testing";
import {KarmaComponent} from './karma.component';

describe('my test', () => {
  it('should be true', () => {
    console.log('test');
    expect(true).toBe(true);
  });
});




describe('KarmaComponent', () => {
  let component: KarmaComponent;
  let fixture: ComponentFixture<KarmaComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [KarmaComponent],
      imports: [BrowserTestingModule]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(KarmaComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
