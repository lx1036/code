
import { OnInit } from '@angular/core';
import {SideNavCollapseStorage} from "../shared.const";

export class SideNavCollapse implements OnInit {
  _collapsed = false;

  constructor(public storage: any) {}

  get collapsed() {
    return this._collapsed;
  }
  set collapsed(value: boolean) {
    this._collapsed = value;
    this.storage.save(SideNavCollapseStorage, value);
  }

  ngOnInit() {
    this._collapsed = this.storage.get(SideNavCollapseStorage) === 'false' ? false : true;
  }
}

