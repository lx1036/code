import {Injectable} from '@angular/core';


@Injectable()
export class StorageService {

  get(key: string) {
    const value = localStorage.getItem(key);
    if (value.length === 0) {
      return this.getCookie(key);
    }
  }

  getCookie(key: string) {
    return null;
  }
}
