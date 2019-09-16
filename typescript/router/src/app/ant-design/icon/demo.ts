import {Component} from '@angular/core';


@Component({
  selector: 'ng-icon-basic',
  template: `
    <div>
      <i ng-icon type="home"></i>
      <i ng-icon type="setting" theme="fill"></i>
      <i ng-icon type="smile" theme="outline"></i>
      <i ng-icon type="sync" spin="true"></i>
      <i ng-icon type="smile" theme="outline" rotate="180"></i>
      <i ng-icon type="loading"></i>
    </div>
  `
})
export class BasicIcon {

}
