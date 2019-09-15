import {Component, NgModule} from '@angular/core';


@Component({
  selector: 'breadcrumb-basic',
  template: `
    <ng-breadcrumb>
      <ng-breadcrumb-item>
        Step1
      </ng-breadcrumb-item>
      <ng-breadcrumb-item>
        <a>Step2</a>
      </ng-breadcrumb-item>
      <ng-breadcrumb-item>
        Step3
      </ng-breadcrumb-item>
    </ng-breadcrumb>
  `
})
export class BasicBreadcrumb {

}

