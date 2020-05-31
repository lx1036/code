

import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {async, ComponentFixture, TestBed} from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {AppConfig} from '@api/backendapi';
import {SharedModule} from 'shared.module';
import {CardComponent} from '../common/components/card/component';
import {AssetsService} from '../common/services/global/assets';
import {ConfigService} from '../common/services/global/config';
import {AboutComponent} from './component';

describe('AboutComponent', () => {
  let component: AboutComponent;
  let fixture: ComponentFixture<AboutComponent>;
  let httpMock: HttpTestingController;
  let configService: ConfigService;
  let element: HTMLElement;

  // set the predefined values
  const copyrightYear = 2019;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule, HttpClientTestingModule, BrowserAnimationsModule],
      declarations: [AboutComponent, CardComponent],
      providers: [AssetsService, ConfigService],
    }).compileComponents();
    httpMock = TestBed.get(HttpTestingController);
    configService = TestBed.get(ConfigService);
  }));

  beforeEach(async(() => {
    // prepare the component
    configService.init();
    fixture = TestBed.createComponent(AboutComponent);
    component = fixture.componentInstance;

    const configRequest = httpMock.expectOne('config');
    const config: AppConfig = {serverTime: new Date().getTime()};
    configRequest.flush(config);

    // set the fixed values
    component.latestCopyrightYear = copyrightYear;

    // grab the HTML element
    element = fixture.debugElement.query(By.css('kd-card')).nativeElement;
  }));

  it('should print current year', async(() => {
    fixture.detectChanges();
    expect(element.textContent).toContain(`2015 - ${copyrightYear}`);
  }));
});
