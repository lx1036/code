/**
 * A dep is an observable that can have multiple
 * directives subscribing to it.
 */
import {Watcher} from './watcher';


let uid = 0;

export class Dep {
  id: number;
  subs: Array<Watcher> = [];

  constructor() {
    this.id = uid++;
  }

  addSub(sub: Watcher) {
    this.subs.push(sub);
  }

  removeSub() {

  }

  notify() {}
}
