

export class Dep {
  constructor() {
    this.subs = [];
  }
  
  addSub(watcher) {
    this.subs.push(watcher);
  }
}
