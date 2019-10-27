import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { CronjobComponent } from './cronjob.component';

describe('CronjobComponent', () => {
  let component: CronjobComponent;
  let fixture: ComponentFixture<CronjobComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ CronjobComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(CronjobComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
