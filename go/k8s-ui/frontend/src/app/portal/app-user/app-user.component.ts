import { Component, OnInit } from '@angular/core';
import {AppUserService} from '../../shared/client/v1/app-user.service';

@Component({
  selector: 'wayne-app-user',
  template: `
    <div class="clr-row">
      <div class="clr-col-lg-12 clr-col-md-12 clr-col-sm-12 clr-col-xs-12">
        <div class="clr-row flex-items-xs-between flex-items-xs-top" style="padding-left: 15px; padding-right: 15px;">
          <h2 class="header-title">{{'USER.APP_USER_LIST' | translate}}</h2>
        </div>
        <create-edit-app-user (create)="createAppUser($event)"></create-edit-app-user>
        <div class="table-search">
          <div class="table-search-left">
            <wayne-filter-box>
              <wayne-checkbox-group>
                <wayne-checkbox></wayne-checkbox>
              </wayne-checkbox-group>
            </wayne-filter-box>
          </div>
        </div>
        <list-app-user></list-app-user>
      </div>
    </div>
  `,
})
export class AppUserComponent implements OnInit {
  constructor(private appUserService: AppUserService, ) {

  }

  ngOnInit() {
  }

}
