import {Injectable} from '@angular/core';


@Injectable()
export class StorageService {

  get(key: string): string {
    const value = localStorage.getItem(key);
    if (value === null) {
      return '';
    }

    if (value.length === 0) {
      return this.getCookie(key);
    }
  }

  getCookie(key: string): string {
    return '';
  }

  save(SideNavCollapseStorage: string, value: boolean) {
    localStorage.setItem(SideNavCollapseStorage, JSON.stringify(value));
  }
}
